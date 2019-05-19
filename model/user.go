package model

import (
	"database/sql"
	"time"
)

// User is a struct for users table
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserController defines an interface for users table
type UserController interface {
	NewUser(name, email string) (*User, error)
	ListUsers() ([]*User, error)
	GetUser(id int) (*User, error)
	UpdateUser(u *User) (*User, error)
	DeleteUser(id int) error
	Migrate() error
}

// NewUserController creates a controller for users table
func NewUserController(db DB) UserController {
	uc := &userController{}

	uc.db = db

	return uc
}

type userController struct {
	db DB
}

var _ UserController = &userController{}

func (uc *userController) NewUser(name, email string) (*User, error) {
	u := &User{
		Name:  name,
		Email: email,
	}

	err := uc.db.
		QueryRow("INSERT INTO users(name, email) VALUES ($1, $2) RETURNING id, created_at, updated_at", name, email).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *userController) ListUsers() ([]*User, error) {
	rows, err := uc.db.Query("SELECT * FROM users")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*User, 0, 16)
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}
func (uc *userController) GetUser(id int) (*User, error) {
	u := &User{}

	err := uc.db.
		QueryRow("SELECT * FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (uc *userController) UpdateUser(u *User) (*User, error) {
	// copied user to return
	ret := *u

	err := uc.db.
		QueryRow("UPDATE users SET name=$1, email=$2 WHERE id=$3 RETURNING created_at, updated_at", u.Name, u.Email, u.ID).
		Scan(&ret.CreatedAt, &ret.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoUser
		}

		return nil, err
	}

	return &ret, nil
}

func (uc *userController) DeleteUser(id int) error {
	_, err := uc.db.Exec("DELETE FROM users WHERE id=$1", id)

	return err
}

func (uc *userController) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(256),
		email VARCHAR(256) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	DROP TRIGGER IF EXISTS update_tri ON users;
	DROP FUNCTION IF EXISTS set_update_time;
	CREATE FUNCTION set_update_time() RETURNS OPAQUE AS '
		BEGIN
	    	new.updated_at := ''now'';
	    	return new;
  		END;
	' LANGUAGE 'plpgsql';
	CREATE TRIGGER update_tri BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE set_update_time();
	`

	if _, err := uc.db.Exec(query); err != nil {
		return err
	}

	return nil
}
