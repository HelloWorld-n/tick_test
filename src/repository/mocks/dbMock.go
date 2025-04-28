package mocks

import "database/sql"

type RepoMock struct {
	DB *sql.DB
}
