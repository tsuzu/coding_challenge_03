package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

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
			log.Fatal("users table migration error", err)
		}
	}

	handler.UserController = uc

	http.ListenAndServe(":80", handler.GetHandler())
}
