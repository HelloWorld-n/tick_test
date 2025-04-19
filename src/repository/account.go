package repository

import (
	"errors"
	"fmt"
	"strings"

	"tick_test/types"
	errDefs "tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AccountRepository interface {
	EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context)
	UserExists(username string) (exists bool, err error)
	ConfirmAccount(username string, password string) (err error)
	FindAllAccounts() (data []types.AccountGetData, err error)
	FindPaginatedAccounts(pageSize int, pageNumber int) (accounts []types.AccountGetData, err error)
	ConfirmNoAdmins() (adminCount int, err error)
	SaveAccount(obj *types.AccountPostData) (err error)
	DeleteAccount(username string) error
	UpdateExistingAccount(username string, obj *types.AccountPatchData) (err error)
	PromoteExistingAccount(obj *types.AccountPatchPromoteData) (err error)
	FindUserRole(username string) (string, error)
}

func validateCredential(cred string, credName string) (err error) {
	if strings.ContainsAny(cred, "\r\n\x00") {
		return fmt.Errorf("%w: credential %v contains invalid newline characters", errDefs.ErrBadRequest, credName)
	}
	if cred != strings.Trim(cred, " \t\u00A0\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007") {
		return fmt.Errorf("%w: credential %v must not start or end with a space or tab", errDefs.ErrBadRequest, credName)
	}
	return
}

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

func confirmPassword(password string, hash string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return
}

func (r *repo) UserExists(username string) (exists bool, err error) {
	query := `SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`
	err = r.DB.Conn.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %w", err)
	}
	return exists, nil
}

func (r *repo) FindPaginatedAccounts(pageSize int, pageNumber int) (accounts []types.AccountGetData, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}

	offset := (pageNumber - 1) * pageSize

	query := `SELECT username, role FROM book ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.DB.Conn.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts = make([]types.AccountGetData, 0)
	for rows.Next() {
		var account types.AccountGetData
		if err := rows.Scan(&account.Username, &account.Role); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func (r *repo) ConfirmAccount(username string, password string) (err error) {
	query := `SELECT password FROM account WHERE $1 = username`

	rows, err := r.DB.Conn.Query(query, username)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			return
		}
		err = confirmPassword(password, hash)
		return
	}

	if err = rows.Err(); err != nil {
		return
	}
	err = errors.New("unable to find user with given username")
	return
}

func (r *repo) FindAllAccounts() (data []types.AccountGetData, err error) {
	query := `
		SELECT 
			username, 
			(SELECT name FROM role WHERE acc.role_id = id)
		FROM account acc;
	`

	rows, err := r.DB.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data = make([]types.AccountGetData, 0)
	for rows.Next() {
		var account types.AccountGetData
		if err := rows.Scan(&account.Username, &account.Role); err != nil {
			return nil, err
		}
		data = append(data, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (r *repo) ConfirmNoAdmins() (adminCount int, err error) {
	adminQuery := `
		SELECT COUNT(*) 
		FROM account a
		JOIN role r ON a.role_id = r.id
		WHERE r.name = 'Admin'
	`
	err = r.DB.Conn.QueryRow(adminQuery).Scan(&adminCount)
	if err != nil {
		err = fmt.Errorf("%w; error checking existing admin accounts: %w", err, errDefs.ErrInternalServerError)
	}
	if adminCount > 0 {
		err = fmt.Errorf("%w: an admin already exists", errDefs.ErrBadRequest)
	}
	return
}

func (r *repo) SaveAccount(obj *types.AccountPostData) (err error) {
	if err = validateCredential(obj.Username, "Username"); err != nil {
		return
	}
	if err = validateCredential(obj.Password, "Password"); err != nil {
		return
	}
	if obj.Password != obj.SamePassword {
		return fmt.Errorf("%w: field `Password` differs from field `SamePassword`", errDefs.ErrBadRequest)
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`
	err = r.DB.Conn.QueryRow(checkQuery, obj.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking user existence: %w", err)
	}
	if exists {
		return fmt.Errorf("%w; user with username %s", errDefs.ErrDoesExist, obj.Username)
	}
	if obj.Role == "Admin" {
		_, err = r.ConfirmNoAdmins()
		if err != nil {
			return
		}
	}

	query := `
		INSERT INTO account (
			username, 
			password, 
			role_id
		) VALUES ($1, $2, (SELECT id FROM role WHERE name = $3));
	`
	hashedPassword, err := hashPassword(obj.Password)
	if err != nil {
		return
	}

	_, err = r.DB.Conn.Exec(query, obj.Username, hashedPassword, obj.Role)

	return
}

func (r *repo) DeleteAccount(username string) error {
	_, err := r.DB.Conn.Exec(`DELETE FROM account WHERE username = $1`, username)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}
	return nil
}

func (r *repo) UpdateExistingAccount(username string, obj *types.AccountPatchData) (err error) {
	// verify valid input
	if err = validateCredential(obj.Username, "Username"); err != nil {
		return
	}
	if err = validateCredential(obj.Password, "Password"); err != nil {
		return
	}
	var count int
	err = r.DB.Conn.QueryRow(`SELECT COUNT(*) FROM account WHERE username = $1`, username).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: no account found with the specified username", errDefs.ErrConflict)
	}
	if obj.Password != obj.SamePassword {
		return fmt.Errorf("%w: field `Password` differs from field `SamePassword`", errDefs.ErrBadRequest)
	}
	if obj.Password != "" && len(obj.Password) < 8 {
		return fmt.Errorf("%w: field `Password` is too short; excepted lenght at least 8", errDefs.ErrBadRequest)
	}

	// apply changes
	if obj.Password != "" {
		hashedPassword, err := hashPassword(obj.Password)
		if err != nil {
			return err
		}
		_, err = r.DB.Conn.Exec(`UPDATE account SET password = $1 WHERE username = $2`, hashedPassword, username)
		if err != nil {
			return err
		}
	}
	if obj.Username != "" && obj.Username != username {
		_, err = r.DB.Conn.Exec(`UPDATE account SET username = $1 WHERE username = $2`, obj.Username, username)
		if err != nil {
			return err
		}
	}
	return
}

func (r *repo) PromoteExistingAccount(obj *types.AccountPatchPromoteData) (err error) {
	// verify valid input
	var count int
	err = r.DB.Conn.QueryRow(`SELECT COUNT(*) FROM account WHERE username = $1`, obj.Username).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: no account found with the specified username", errDefs.ErrConflict)
	}

	// apply changes
	if obj.Role != "" {
		_, err = r.DB.Conn.Exec(`UPDATE account SET role_id = (SELECT id FROM role WHERE name = $1) WHERE username = $2`, obj.Role, obj.Username)
		if err != nil {
			return err
		}
	}
	return
}

func (r *repo) FindUserRole(username string) (string, error) {
	var role string
	query := `
		SELECT r.name 
		FROM account a 
		JOIN role r ON a.role_id = r.id 
		WHERE a.username = $1
	`
	err := r.DB.Conn.QueryRow(query, username).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (r *repo) doPostgresPreparationForAccount() {
	if r.DB.Conn != nil {
		result, err := r.DB.Conn.Exec(`
			CREATE TABLE IF NOT EXISTS account (
				username varchar(100) PRIMARY KEY,
				password varchar(500) NOT NULL
			);
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			CREATE TABLE IF NOT EXISTS role (
				id SERIAL PRIMARY KEY,
				name TEXT UNIQUE NOT NULL
			);
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			INSERT INTO role (name) VALUES
				('User'),
				('BookKeeper'),
				('Admin')
			ON CONFLICT (name) DO NOTHING;
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account ADD COLUMN IF NOT EXISTS role_id INT REFERENCES role(id);
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			UPDATE account
			SET role_id = (SELECT id FROM role WHERE name = 'User')
			WHERE role_id IS NULL;
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account
			ALTER COLUMN role_id SET NOT NULL;
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account ADD COLUMN IF NOT EXISTS id SERIAL;
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account DROP CONSTRAINT IF EXISTS account_pkey;
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account ADD PRIMARY KEY (id);
		`)
		fmt.Println(result, err)
		result, err = r.DB.Conn.Exec(`
			ALTER TABLE account ADD CONSTRAINT unique_username UNIQUE(username);
		`)
		fmt.Println(result, err)
	}
}
