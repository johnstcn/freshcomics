package main

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/johnstcn/freshcomics/internal/store"
	"github.com/johnstcn/freshcomics/internal/web"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func root() *cobra.Command {
	var (
		host string
		port int
		dsn  string
	)
	rootCmd := &cobra.Command{
		Use:   "freshcomics",
		Short: "Keep up to date on your favourite webcomics",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := sqlx.Connect("postgres", dsn)
			if err != nil {
				return fmt.Errorf("connect to db: %w", err)
			}

			store, err := store.NewPGStore(conn)
			if err != nil {
				return fmt.Errorf("init store: %w", err)
			}

			listenAddress := fmt.Sprintf("%s:%d", host, port)
			mux := http.NewServeMux()

			log := slog.New(slog.NewTextHandler(cmd.OutOrStdout()))
			fe := web.New(web.Deps{
				Mux:    mux,
				Store:  store,
				Logger: log,
			})
			slog.Info("listen", "host", host, "port", port)
			return http.ListenAndServe(listenAddress, fe)
		},
	}

	return rootCmd
}
