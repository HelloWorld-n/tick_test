package mocks

import (
	"tick_test/types"

	"github.com/gin-gonic/gin"
)

type AccountRepositoryMock struct {
	EnsureDatabaseIsOkFn     func(func(*gin.Context)) func(*gin.Context)
	UserExistsFn             func(string) (bool, error)
	ConfirmAccountFn         func(string, string) error
	FindAllAccountsFn        func() ([]types.AccountGetData, error)
	ConfirmNoAdminsFn        func() (int, error)
	SaveAccountFn            func(*types.AccountPostData) error
	DeleteAccountFn          func(string) error
	UpdateExistingAccountFn  func(string, *types.AccountPatchData) (int64, error)
	PromoteExistingAccountFn func(*types.AccountPatchPromoteData) error
	FindUserRoleFn           func(string) (string, error)
	FindPaginatedAccountsFn  func(page, size int) ([]types.AccountGetData, error)
}

func (arm *AccountRepositoryMock) EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
	return arm.EnsureDatabaseIsOkFn(fn)
}

func (arm *AccountRepositoryMock) UserExists(username string) (exists bool, err error) {
	return arm.UserExistsFn(username)
}

func (arm *AccountRepositoryMock) ConfirmAccount(username string, password string) (err error) {
	return arm.ConfirmAccountFn(username, password)
}

func (arm *AccountRepositoryMock) FindAllAccounts() (data []types.AccountGetData, err error) {
	return arm.FindAllAccountsFn()
}

func (arm *AccountRepositoryMock) ConfirmNoAdmins() (adminCount int, err error) {
	return arm.ConfirmNoAdminsFn()
}

func (arm *AccountRepositoryMock) SaveAccount(obj *types.AccountPostData) (err error) {
	return arm.SaveAccountFn(obj)
}

func (arm *AccountRepositoryMock) DeleteAccount(username string) error {
	return arm.DeleteAccountFn(username)
}

func (arm *AccountRepositoryMock) UpdateExistingAccount(username string, obj *types.AccountPatchData) (int64, error) {
    return arm.UpdateExistingAccountFn(username, obj)
}

func (arm *AccountRepositoryMock) PromoteExistingAccount(obj *types.AccountPatchPromoteData) (err error) {
	return arm.PromoteExistingAccountFn(obj)
}

func (arm *AccountRepositoryMock) FindUserRole(username string) (string, error) {
	return arm.FindUserRoleFn(username)
}

func (m *AccountRepositoryMock) FindPaginatedAccounts(page, size int) ([]types.AccountGetData, error) {
	if m.FindPaginatedAccountsFn != nil {
		return m.FindPaginatedAccountsFn(page, size)
	}
	return nil, nil
}
