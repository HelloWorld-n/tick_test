package mocks

import (
	"tick_test/types"
	"tick_test/utils/jwt"

	"github.com/gin-gonic/gin"
)

type AccountRepositoryMock struct {
	EnsureDatabaseIsOkFn      func(func(*gin.Context)) func(*gin.Context)
	UserExistsFn              func(string) (bool, error)
	ConfirmAccountFn          func(string, string) error
	ConfirmAccountJwtFn       func(string, string) (string, error)
	FindAccountIdByUsernameFn func(username string) (int64, error)
	FindAllAccountsFn         func() ([]types.AccountGetData, error)
	FindPaginatedAccountsFn   func(pageSize, pageNumber int) ([]types.AccountGetData, error)
	ConfirmNoAdminsFn         func() (int, error)
	SaveAccountFn             func(*types.AccountPostData) error
	DeleteAccountFn           func(string) error
	UpdateExistingAccountFn   func(string, *types.AccountPatchData) (int64, error)
	PromoteExistingAccountFn  func(*types.AccountPatchPromoteData) error
	FindUserRoleFn            func(string) (types.Role, error)
	ValidateTokenFn           func(string) (jwt.Claims, error)
	GenerateTokenForUserFn    func(string) (string, error)
	IsAdminFn                 func(string) (bool, error)
}

func (arm *AccountRepositoryMock) EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
	return arm.EnsureDatabaseIsOkFn(fn)
}

func (arm *AccountRepositoryMock) UserExists(username string) (bool, error) {
	return arm.UserExistsFn(username)
}

func (arm *AccountRepositoryMock) ConfirmAccount(username, password string) error {
	return arm.ConfirmAccountFn(username, password)
}

func (arm *AccountRepositoryMock) ConfirmAccountJwt(username, password string) (string, error) {
	return arm.ConfirmAccountJwtFn(username, password)
}

func (arm *AccountRepositoryMock) FindAccountIdByUsername(username string) (int64, error) {
	return arm.FindAccountIdByUsernameFn(username)
}

func (arm *AccountRepositoryMock) FindAllAccounts() ([]types.AccountGetData, error) {
	return arm.FindAllAccountsFn()
}

func (arm *AccountRepositoryMock) FindPaginatedAccounts(pageSize, pageNumber int) ([]types.AccountGetData, error) {
	if arm.FindPaginatedAccountsFn != nil {
		return arm.FindPaginatedAccountsFn(pageSize, pageNumber)
	}
	return nil, nil
}

func (arm *AccountRepositoryMock) ConfirmNoAdmins() (int, error) {
	return arm.ConfirmNoAdminsFn()
}

func (arm *AccountRepositoryMock) SaveAccount(obj *types.AccountPostData) error {
	return arm.SaveAccountFn(obj)
}

func (arm *AccountRepositoryMock) DeleteAccount(username string) error {
	return arm.DeleteAccountFn(username)
}

func (arm *AccountRepositoryMock) UpdateExistingAccount(username string, obj *types.AccountPatchData) (int64, error) {
	return arm.UpdateExistingAccountFn(username, obj)
}

func (arm *AccountRepositoryMock) PromoteExistingAccount(obj *types.AccountPatchPromoteData) error {
	return arm.PromoteExistingAccountFn(obj)
}

func (arm *AccountRepositoryMock) FindUserRole(username string) (types.Role, error) {
	return arm.FindUserRoleFn(username)
}

func (arm *AccountRepositoryMock) ValidateToken(token string) (jwt.Claims, error) {
	return arm.ValidateTokenFn(token)
}

func (arm *AccountRepositoryMock) GenerateTokenForUser(username string) (string, error) {
	return arm.GenerateTokenForUserFn(username)
}

func (arm *AccountRepositoryMock) IsAdmin(token string) (bool, error) {
	return arm.IsAdminFn(token)
}
