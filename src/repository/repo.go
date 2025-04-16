package repository

import "database/sql"

type repo struct {
	DB *Database
}

func NewRepo(db *Database) *repo {
	return &repo{DB: db}
}

type Repository interface {
	AccountRepository
	BookRepository
	ManipulatorRepository
	MessageRepository
	DoPostgresPreparation() (db *sql.DB, err error)
}
