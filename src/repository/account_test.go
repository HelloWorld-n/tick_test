package repository_test

import (
	"errors"
	"regexp"
	"testing"
	"tick_test/repository"
	"tick_test/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserExists(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		mockResult    bool
		mockError     error
		expectedExist bool
		expectError   bool
	}{
		{
			name:          "User exists",
			username:      "john",
			mockResult:    true,
			expectedExist: true,
			expectError:   false,
		},
		{
			name:          "User doesn't exist",
			username:      "john",
			mockResult:    false,
			expectedExist: false,
			expectError:   false,
		},
		{
			name:        "Database error",
			username:    "john",
			mockError:   errors.New("db error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`)
			expect := mock.ExpectQuery(query).WithArgs(tt.username)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnRows(
					sqlmock.NewRows([]string{"exists"}).AddRow(tt.mockResult),
				)
			}

			exists, err := r.UserExists(tt.username)

			if tt.expectError {
				require.ErrorContains(t, err, "error checking user existence")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedExist, exists)
			}
		})
	}
}

func TestConfirmAccount(t *testing.T) {
	validPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		username       string
		inputPassword  string
		dbPassword     string
		dbRowsReturned bool
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "Valid credentials",
			username:       "john",
			inputPassword:  validPassword,
			dbPassword:     string(hashedPassword),
			dbRowsReturned: true,
			expectError:    false,
		},
		{
			name:           "Invalid password",
			username:       "john",
			inputPassword:  "wrongpassword",
			dbPassword:     string(hashedPassword),
			dbRowsReturned: true,
			expectError:    true,
		},
		{
			name:           "User not found",
			username:       "john",
			inputPassword:  "password",
			dbRowsReturned: false,
			expectError:    true,
			expectedErrMsg: "unable to find user with given username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`SELECT password FROM account WHERE $1 = username`)
			expect := mock.ExpectQuery(query).WithArgs(tt.username)

			if tt.dbRowsReturned {
				expect.WillReturnRows(
					sqlmock.NewRows([]string{"password"}).AddRow(tt.dbPassword),
				)
			} else {
				expect.WillReturnRows(sqlmock.NewRows([]string{"password"}))
			}

			err := r.ConfirmAccount(tt.username, tt.inputPassword)

			if tt.expectError {
				if tt.expectedErrMsg != "" {
					require.EqualError(t, err, tt.expectedErrMsg)
				} else {
					require.Error(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFindPaginatedAccounts(t *testing.T) {
	tests := []struct {
		name              string
		limit             int
		page              int
		mockRows          *sqlmock.Rows
		mockError         error
		expectError       bool
		expectedLen       int
		expectedErrorText string
	}{
		{
			name:  "Success with multiple accounts",
			limit: 10,
			page:  1,
			mockRows: sqlmock.NewRows([]string{"username", "name"}).
				AddRow("john", "User").
				AddRow("admin", "Admin"),
			expectedLen: 2,
		},
		{
			name:        "Success with no accounts",
			limit:       10,
			page:        2,
			mockRows:    sqlmock.NewRows([]string{"username", "name"}),
			expectedLen: 0,
		},
		{
			name:        "Wrong page",
			limit:       10,
			page:        0,
			mockError:   errors.New("wrong page"),
			expectError: true,
		},
		{
			name:        "Wrong limit",
			limit:       10,
			page:        0,
			mockError:   errors.New("wrong limit"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(
				`SELECT username, (SELECT name FROM role WHERE acc.role_id = id) FROM account acc ORDER BY id LIMIT $1 OFFSET $2`,
			)

			offset := (tt.page - 1) * tt.limit
			expect := mock.ExpectQuery(query).WithArgs(tt.limit, offset)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnRows(tt.mockRows)
			}

			result, err := r.FindPaginatedAccounts(tt.limit, tt.page)

			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedErrorText)
			} else {
				require.NoError(t, err)
				require.Len(t, result, tt.expectedLen)
			}
		})
	}
}

func TestFindAllAccounts(t *testing.T) {
	rMock, mock := setupMock(t)
	r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
	defer r.DB.Conn.Close()

	rows := sqlmock.NewRows([]string{"username", "name"}).
		AddRow("john", "User").
		AddRow("admin", "Admin")

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT username, (SELECT name FROM role WHERE acc.role_id = id) FROM account acc;`,
	)).WillReturnRows(rows)

	result, err := r.FindAllAccounts()
	require.NoError(t, err)
	require.Len(t, result, 2)
}

func TestConfirmNoAdmins(t *testing.T) {
	tests := []struct {
		name           string
		adminCount     int
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "No admins exist",
			adminCount:  0,
			expectError: false,
		},
		{
			name:           "Admin exists",
			adminCount:     1,
			expectError:    true,
			expectedErrMsg: "an admin already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(
				`SELECT COUNT(*) FROM account a JOIN role r ON a.role_id = r.id WHERE r.name = 'Admin'`,
			)

			mock.ExpectQuery(query).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.adminCount))

			count, err := r.ConfirmNoAdmins()

			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.adminCount, count)
			}
		})
	}
}

func TestSaveAccount(t *testing.T) {
	validAccount := &types.AccountPostData{
		Username:     "john",
		Password:     "password123",
		SamePassword: "password123",
		Role:         "User",
	}

	tests := []struct {
		name              string
		account           *types.AccountPostData
		userExists        bool
		expectInsert      bool
		expectError       bool
		expectedErrorText string
	}{
		{
			name:         "Success",
			account:      validAccount,
			userExists:   false,
			expectInsert: true,
			expectError:  false,
		},
		{
			name:              "Existing user",
			account:           validAccount,
			userExists:        true,
			expectInsert:      false,
			expectError:       true,
			expectedErrorText: "user with username john",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			// Expect user existence check
			mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`,
			)).
				WithArgs(tt.account.Username).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(tt.userExists))

			if tt.expectInsert {
				// Expect account insert
				mock.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO account ( username, password, role_id ) VALUES ($1, $2, (SELECT id FROM role WHERE name = $3));`,
				)).
					WithArgs(tt.account.Username, sqlmock.AnyArg(), tt.account.Role).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := r.SaveAccount(tt.account)

			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedErrorText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	tests := []struct {
		name              string
		username          string
		mockError         error
		expectError       bool
		expectedErrorText string
	}{
		{
			name:        "Success",
			username:    "john",
			expectError: false,
		},
		{
			name:              "Database error",
			username:          "john",
			mockError:         errors.New("db error"),
			expectError:       true,
			expectedErrorText: "error deleting account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`DELETE FROM account WHERE username = $1`)
			exec := mock.ExpectExec(query).WithArgs(tt.username)

			if tt.mockError != nil {
				exec.WillReturnError(tt.mockError)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
			}

			err := r.DeleteAccount(tt.username)

			if tt.expectError {
				require.ErrorContains(t, err, tt.expectedErrorText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPromoteExistingAccount(t *testing.T) {
	validPromotion := &types.AccountPatchPromoteData{
		Username: "john",
		Role:     "Admin",
	}

	t.Run("Success", func(t *testing.T) {
		rMock, mock := setupMock(t)
		r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
		defer r.DB.Conn.Close()

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT COUNT(*) FROM account WHERE username = $1`,
		)).
			WithArgs("john").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE account SET role_id = (SELECT id FROM role WHERE name = $1) WHERE username = $2`,
		)).
			WithArgs("Admin", "john").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := r.PromoteExistingAccount(validPromotion)
		require.NoError(t, err)
	})
}

func TestFindUserRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rMock, mock := setupMock(t)
		r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
		defer r.DB.Conn.Close()

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT r.name FROM account a JOIN role r ON a.role_id = r.id WHERE a.username = $1`,
		)).
			WithArgs("john").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Admin"))

		role, err := r.FindUserRole("john")
		require.NoError(t, err)
		require.Equal(t, types.Role("Admin"), role)
	})
}
