package main

import (
	"backend/database"
	"backend/internal/api"
	"backend/internal/discord"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// DB接続
	db, err := database.NewConnection()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	// DI
	whitelistRepo := repository.NewWhitelistRepository(db)
	whitelistService := service.NewWhitelistService(whitelistRepo)
	whitelistHandler := api.NewWhitelistHandler(whitelistService)

	healthRepo := repository.NewHealthRepository(db)
	healthSevice := service.NewHealthService(healthRepo)
	healthHandler := api.NewHealthHandler(healthSevice)

	// Echo インスタンスを作成
	e := echo.New()

	// ミドルウェア
	// 起動時のASCIIバナーを消す
	e.HideBanner = true
	// /users/ のような末尾スラッシュをリクエスト内で剥がして /users に書き換える
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(
		middleware.Recover(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.CORS(),
	)

	// ルート設定
	api.SetupRoutes(e, healthHandler, whitelistHandler)

	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ---- server with timeouts ----
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      e,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- 起動ウォームアップ：依存OKならready ON ----
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		// まずflagをONにしてから疎通を見る。
		healthSevice.MarkReady()
		if !healthSevice.Ready(ctx) {
			// 依存がまだならflagをOFF
			healthSevice.MarkNotReady()
		}
	}()

	// ========= Discord セッション準備 =========
	discordToken := os.Getenv("DISCORD_TOKEN")
	discordAppID := os.Getenv("DISCORD_APP_ID")
	discordGuildID := os.Getenv("DISCORD_GUILD_ID") // dev中は Guild 指定推奨

	var dSession discord.Session
	if discordToken != "" {
		s, err := discord.NewSession(discordToken)
		if err != nil {
			e.Logger.Fatal("failed to init discord session: ", err)
		}
		dSession = s
	} else {
		e.Logger.Warn("DISCORD_TOKEN not set: discord bot disabled")
	}

	// ---- server start & wait for signal ----
	// サーバ起動結果（エラー）を受け取るためのチャネルを用意する（バッファ1で送信ブロックを避ける）
	// Discord分も見たいので容量2に
	errCh := make(chan error, 2)

	// サーバを別ゴルーチンで起動する
	// HTTPサーバ起動
	go func() {
		if err := e.StartServer(srv); err != nil {
			errCh <- err
		}
	}()

	// Discord起動
	if dSession != nil {
		router := discord.NewRouter(whitelistService)
		dSession.AddHandler(router.HandleInteraction)

		go func() {
			ctxStart, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := dSession.Start(ctxStart); err != nil {
				errCh <- fmt.Errorf("discord start: %w", err)
				return
			}

			ctxCmd, cancelCmd := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelCmd()

			if err := dSession.RegisterCommands(ctxCmd, discordAppID, discordGuildID); err != nil {
				errCh <- fmt.Errorf("discord register commands: %w", err)
				return
			}

			fmt.Printf("startup complete: http=:%s, discord=online\n", port)
		}()
	} else {
		fmt.Printf("startup complete: http=:%s, discord=disabled\n", port)
	}

	// OSシグナル（Ctrl+C の SIGINT と SIGTERM）を受けると自動で Done になるコンテキストを作る
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 「シグナルでの終了要求」か「サーバ起動側のエラー」のどちらが先かを競合待ちする
	select {
	case <-ctx.Done():
		// シグナルを受けたのでシャットダウンへ進む。 ログを出す。
		e.Logger.Info("Server is shutting down...")
	case err := <-errCh:
		// サーバ起動側が先に戻った（起動失敗 or 正常終了）
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			// それ以外はポート競合などの致命的な起動失敗とみなして落とす
			e.Logger.Fatal(err)
		}
	}

	// ---- graceful shutdown ----
	// まずreadyを落としてロードバランサから外れる（ドレイン）
	healthSevice.MarkNotReady()
	// 猶予時間を設定（ここでは10秒）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 新規受付を止める
	if err := e.Shutdown(shutdownCtx); err != nil {
		// 猶予内に閉じられない等で失敗した場合はログに残し
		e.Logger.Error("graceful shutdown failed, forcing close:", err)
		// 最終手段として強制クローズ（未完リクエストはエラーになる前提）
		if cerr := e.Close(); cerr != nil {
			e.Logger.Error(cerr)
		}
	}

	// Discordを閉じる（WebSocket切断）
	if err := dSession.Close(); err != nil {
		e.Logger.Error("discord close:", err)
	}

	// DBはここで閉じる（全リクエスト完了後）
	if derr := db.Close(); derr != nil {
		e.Logger.Error("db close:", derr)
	}

	e.Logger.Info("Server stopped")

}
