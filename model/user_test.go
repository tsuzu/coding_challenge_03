package model_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/cs3238-tsuzu/coding_challenge_03/model"
	_ "github.com/lib/pq"
)

const migration = `
DROP TRIGGER IF EXISTS update_tri ON users;
DROP FUNCTION IF EXISTS set_update_time;
DROP TABLE IF EXISTS users;
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name VARCHAR(256) NOT NULL,
	email VARCHAR(256) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE FUNCTION set_update_time() RETURNS OPAQUE AS '
	BEGIN
		new.updated_at := ''now'';
		return new;
	  END;
' LANGUAGE 'plpgsql';
CREATE TRIGGER update_tri BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE set_update_time();
`

func initDB(t *testing.T) (*sql.DB, model.UserController) {
	t.Helper()

	dsn := os.Getenv("POSTGRES_DSN")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec(migration); err != nil {
		t.Fatal("migration error", err)
	}

	return db, model.NewUserController(db)
}

func compareUser(t *testing.T, real *model.User, expected *model.User) {
	t.Helper()

	if expected.ID != -1 && real.ID != expected.ID {
		t.Errorf("id does not match (expected: %v, actual: %v)", expected.ID, real.ID)
	}

	if real.Name != expected.Name {
		t.Errorf("name does not match (expected: %v, actual: %v)", expected.Name, real.Name)
	}

	if real.Email != expected.Email {
		t.Errorf("email does not match (expected: %v, actual: %v)", expected.Email, real.Email)
	}
}

func checkTime(t *testing.T, before, after, target time.Time) {
	t.Helper()

	if target.Before(before.Add(-1*time.Second)) || target.After(after.Add(1*time.Second)) {
		t.Fatalf("time is invalid(should be in (%v, %v), but got %v)", before, after, target)
	}
}
func TestNewUser(t *testing.T) {
	before := time.Now()
	db, uc := initDB(t)

	param := &model.User{
		ID:    -1,
		Name:  "name",
		Email: "hoge@example.com",
	}

	ret, err := uc.NewUser(param.Name, param.Email)

	if err != nil {
		t.Fatal("new user error ", err)
	}

	compareUser(t, ret, param)

	after := time.Now()

	checkTime(t, before, after, ret.CreatedAt)
	checkTime(t, before, after, ret.UpdatedAt)

	rows, err := db.Query("SELECT id, name, email, created_at, updated_at FROM users")

	if err != nil {
		t.Fatal("select from users error ", err)
	}
	defer rows.Close()

	rows.Next()

	var user model.User
	if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		t.Fatal("row scan error ", err)
	}

	if rows.Next() {
		t.Error("count(*) should be 1")
	}

	compareUser(t, &user, ret)
}

func TestListUsers(t *testing.T) {
	before := time.Now()
	_, uc := initDB(t)

	params := []*model.User{
		{
			ID:    -1,
			Name:  "name",
			Email: "hoge@example.com",
		},
		{
			ID:    -1,
			Name:  "name2",
			Email: "hoge2@example.com",
		},
	}

	for _, p := range params {
		if _, err := uc.NewUser(p.Name, p.Email); err != nil {
			t.Fatal("new user error ", err)
		}
	}

	after := time.Now()

	users, err := uc.ListUsers()

	if err != nil {
		t.Fatal("list user error ", err)
	}

	for i, u := range users {
		compareUser(t, u, params[i])

		checkTime(t, before, after, u.CreatedAt)
		checkTime(t, before, after, u.UpdatedAt)
	}
}

func TestGetUser(t *testing.T) {
	before := time.Now()
	_, uc := initDB(t)

	param := &model.User{
		ID:    -1,
		Name:  "name",
		Email: "hoge@example.com",
	}
	user, err := uc.NewUser(param.Name, param.Email)

	if err != nil {
		t.Fatal("new user error ", err)
	}

	after := time.Now()

	ret, err := uc.GetUser(user.ID)

	if err != nil {
		t.Fatal("list user error ", err)
	}

	compareUser(t, ret, user)

	checkTime(t, before, after, ret.CreatedAt)
	checkTime(t, before, after, ret.UpdatedAt)
}

func TestUpdateUser(t *testing.T) {
	beforeCreated := time.Now()
	_, uc := initDB(t)

	param := &model.User{
		ID:    -1,
		Name:  "name",
		Email: "hoge@example.com",
	}
	user, err := uc.NewUser(param.Name, param.Email)

	if err != nil {
		t.Fatal("new user error ", err)
	}

	afterCreated := time.Now()

	time.Sleep(3 * time.Second)

	beforeUpdated := time.Now()

	user.Name = "name2"
	user.Email = "hoge2@example.com"

	ret, err := uc.UpdateUser(user)

	if err != nil {
		t.Fatal("update error ", err)
	}
	afterUpdated := time.Now()

	compareUser(t, ret, user)

	checkTime(t, beforeCreated, afterCreated, ret.CreatedAt)
	checkTime(t, beforeUpdated, afterUpdated, ret.UpdatedAt)

	ret, err = uc.GetUser(user.ID)

	if err != nil {
		t.Fatal("get user error ", err)
	}

	compareUser(t, ret, user)

	checkTime(t, beforeCreated, afterCreated, ret.CreatedAt)
	checkTime(t, beforeUpdated, afterUpdated, ret.UpdatedAt)
}

func TestMigrate(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	uc := model.NewUserController(db)

	if err := uc.Migrate(); err != nil {
		t.Fatal("migration error", err)
	}

	// check idempotency
	if err := uc.Migrate(); err != nil {
		t.Fatal("migration for checking idempotency error", err)
	}
}
