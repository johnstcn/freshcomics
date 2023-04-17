package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"

	"github.com/johnstcn/freshcomics/internal/api"
	"github.com/johnstcn/freshcomics/internal/store"
)

func main() {
	var (
		host string
		port int
		dsn  string
		log  = slog.New(slog.NewTextHandler(os.Stdout))
	)

	flag.StringVar(&host, "host", "0.0.0.0", "listen on this host")
	if val, ok := os.LookupEnv("FRESHCOMICS_HOST"); ok {
		host = val
	}

	flag.IntVar(&port, "port", 8000, "listen on this port")
	if val, ok := os.LookupEnv("FRESHCOMICS_PORT"); ok {
		if intVal, err := strconv.Atoi(val); err != nil {
			log.Error("invalid port env", "val", val)
			os.Exit(1)
		} else {
			port = intVal
		}
	}

	flag.StringVar(&dsn, "dsn", "postgresql://localhost:5432/freshcomics?user=freshcomics&password=freshcomics", "postgresql connection string")
	if val, ok := os.LookupEnv("FRESHCOMICS_DB"); ok {
		dsn = val
	}

	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Error("connect to db", "err", err)
		os.Exit(1)
	}

	store, err := store.NewPGStore(conn)
	if err != nil {
		log.Error("init store", "err", err)
		os.Exit(1)
	}

	listenAddress := fmt.Sprintf("%s:%d", host, port)
	mux := http.NewServeMux()
	srv := api.New(api.Deps{
		Mux:    mux,
		Store:  store,
		Logger: log,
	})

	slog.Info("listen", "host", host, "port", port)
	if err := http.ListenAndServe(listenAddress, srv); err != nil {
		log.Error("listen and serve", "err", err)
	}
}
