package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func NewConnection() (*sql.DB, error) {
	host := getEnvWithDefault("DB_HOST", "localhost")
	port := getEnvWithDefault("DB_PORT", "5432")
	user := mustEnv("DB_USER")
	password := mustEnv("DB_PASSWORD")
	dbname := getEnvWithDefault("DB_NAME", "chatclub")

	// DSN生成
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&timezone=UTC",
		user, password, host, port, dbname)

	// 接続ハンドル作成
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// 到達性チェック 指数バックオフでPingを繰り返す
	// backoffを100msで初期化
	for backoff := 100 * time.Millisecond; ; {
		// 2秒のタイムアウトを設定
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := db.PingContext(ctx)
		cancel()

		// Ping成功ならループを抜けて終了
		if err == nil {
			break
		}
		// 一定以上リトライしても成功しなければ諦めてエラーを返す
		if backoff > 2*time.Second {
			return nil, err
		}
		// 次のPingまで指定時間だけ待つ
		time.Sleep(backoff)
		// 待機時間を2倍にして再試行間隔を伸ばす（指数バックオフ）
		backoff *= 2
	}

	return db, nil
}

// 取得した環境変数が空文字ならdefaultValueを返す
func getEnvWithDefault(key, defaultValue string) string {
	// 環境変数を取得
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 必須の環境変数が未設定なら即座にプロセスを止める
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		// プログラムを強制的に終了
		panic("missing required env: " + k)
	}
	return v
}
