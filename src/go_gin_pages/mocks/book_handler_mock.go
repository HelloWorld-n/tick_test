package mocks

import (
	"tick_test/types"

	"github.com/gin-gonic/gin"
)

type BookRepositoryMock struct {
	EnsureDatabaseIsOKFn func(func(*gin.Context)) func(*gin.Context)
	FindAllBooksFn       func() ([]types.Book, error)
	FindPaginatedBooksFn func(int, int) ([]types.Book, error)
	FindBookByCodeFn     func(string) (types.Book, error)
	CreateBookFn         func(*types.Book) error
	UpdateBookByCodeFn   func(string, types.Book) (types.Book, error)
	RemoveBookByCodeFn   func(string) (int64, error)
}

func (brm *BookRepositoryMock) EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
	return brm.EnsureDatabaseIsOKFn(fn)
}

func (brm *BookRepositoryMock) FindAllBooks() (books []types.Book, err error) {
	return brm.FindAllBooksFn()
}

func (brm *BookRepositoryMock) FindPaginatedBooks(pageSize int, pageNumber int) (books []types.Book, err error) {
	return brm.FindPaginatedBooksFn(pageSize, pageNumber)
}

func (brm *BookRepositoryMock) FindBookByCode(code string) (book types.Book, err error) {
	return brm.FindBookByCodeFn(code)
}

func (brm *BookRepositoryMock) CreateBook(book *types.Book) (err error) {
	return brm.CreateBookFn(book)
}

func (brm *BookRepositoryMock) UpdateBookByCode(code string, updates types.Book) (book types.Book, err error) {
	return brm.UpdateBookByCodeFn(code, updates)
}

func (brm *BookRepositoryMock) RemoveBookByCode(code string) (n int64, err error) {
	return brm.RemoveBookByCodeFn(code)
}
