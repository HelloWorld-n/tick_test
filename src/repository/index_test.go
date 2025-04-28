package repository_test

import (
	"testing"
	"tick_test/repository/mocks"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func setupMock(t *testing.T) (*mocks.RepoMock, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &mocks.RepoMock{DB: db}, mock
}
