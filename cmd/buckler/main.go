package main

import (
	"backend/internal/buckler"
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	var (
		login = flag.Bool("login", false, "run login flow")
		fetch = flag.Bool("fetch", false, "fetch battlelog")
		sid   = flag.String("sid", "", "Buckler short_id (sid)")
		page  = flag.Int("page", 1, "page number")
	)
	flag.Parse()

	if !*login && !*fetch {
		fmt.Println("usage: buckler --login | --fetch --sid <sid> [--page n]")
		os.Exit(2)
	}

	cfg, err := buckler.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	client, err := buckler.NewClient(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "client error: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if *login {
		if err := client.Login(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("login ok")
	}

	if *fetch {
		if *sid == "" {
			fmt.Fprintln(os.Stderr, "--sid required")
			os.Exit(2)
		}
		res, err := client.FetchCustomBattlelog(ctx, *sid, *page)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("fetch ok: page=%d total_page=%d sid=%d items=%d\n",
			res.PageProps.CurrentPage,
			res.PageProps.TotalPage,
			res.PageProps.SID,
			len(res.PageProps.ReplayList),
		)
	}
}
