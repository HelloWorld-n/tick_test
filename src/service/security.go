package service

import (
	"fmt"
	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"
	"tick_test/utils/jwt"

	"github.com/gin-gonic/gin"
)

var ErrUnauthorized = errDefs.ErrUnauthorized

type SecurityService struct {
	AccRepo repository.AccountRepository
}

func (ss *SecurityService) CheckRole(c *gin.Context, role types.Role) (types.Account, bool, error) {
	acc := types.Account{}
	claims, err := jwt.ValidateToken(c.Request.Header.Get("User-Token"))
	if err != nil {
		return acc, false, fmt.Errorf("%w: %v", errDefs.ErrUnauthorized, err)
	}
	acc.Id, err = ss.AccRepo.FindAccountIdByUsername(claims.Username)
	if err != nil {
		return acc, false, fmt.Errorf("%w: %v", errDefs.ErrUnauthorized, err)
	}
	acc.Username = claims.Username
	acc.Role = types.Role(claims.Role)
	if role == types.Role(acc.Role) {
		return acc, true, nil
	}
	return acc, true, nil
}
