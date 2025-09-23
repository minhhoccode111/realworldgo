package repository

import (
	"database/sql"

	"github.com/minhhoccode111/realworldgo/internal/entity"
)

type UserPostgres struct {
	db *sql.DB
}

func NewUserPostgres(db *sql.DB) *UserPostgres {
	return &UserPostgres{
		db: db,
	}
}

func (up *UserPostgres) GetByID(id int) (*entity.User, error) {
	row := up.db.QueryRow("SELECT id, name FROM users WHERE id=$1", id)
	u := entity.User{}
	err := row.Scan(&u.ID, &u.Username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
