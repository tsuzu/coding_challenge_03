package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cs3238-tsuzu/coding_challenge_03/handler"
	"github.com/cs3238-tsuzu/coding_challenge_03/model"
	_ "github.com/lib/pq"
)

var (
	migrate = flag.Bool("migrate", false, "execute migration")
	dsn     = flag.String("db", "", "data source name")
	help    = flag.Bool("help", false, "Show usage")
)

func main() {
	flag.Parse()
	dsn := *dsn

	if *help {
		flag.Usage()

		return
	}

	dsnEnv := os.Getenv("POSTGRES_DSN")

	if len(dsn) == 0 {
		dsn = dsnEnv
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	handler := handler.NewHandler(db)

	uc := model.NewUserController(db)

	if *migrate {
		if err := uc.Migrate(); err != nil {
			log.Fatal("users table migration error: ", err)
		}
	}

	handler.UserController = uc

	server := http.Server{
		Addr:    ":80",
		Handler: handler.GetHandler(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("listen and server error: ", err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-sig

	ctx, canceler := context.WithTimeout(context.Background(), 5*time.Second)
	defer canceler()

	if err := server.Shutdown(ctx); err != nil {
		log.Print(err)
	}
}
