package admin

import (
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByEmail(email string) (*User, error) {
	var u User
	err := r.db.Get(&u, "SELECT * FROM admins WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
