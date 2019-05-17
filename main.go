package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cs3238-tsuzu/coding_challenge_03/handler"
	"github.com/cs3238-tsuzu/coding_challenge_03/model"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {

	connStr := os.Getenv("POSTGRES")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	handler := handler.NewHandler(db)

	uc := model.NewUserController(db)

	if err := uc.Migrate(); err != nil {
		log.Fatal("users table migration error", err)
	}

	handler.UserController = uc

	http.ListenAndServe(":80", handler.GetHandler())
}
