package types

type Role string

const (
	AdminRole      Role = "Admin"
	BookKeeperRole Role = "BookKeeper"
	UserRole       Role = "User"
)

type Account struct {
	Id       int64  `json:"id"`
	Username string `json:"username" binding:"required"`
	Role     Role   `json:"role" binding:"required"`
}

type AccountPostData struct {
	Username     string `json:"username" binding:"gt=4"`
	Password     string `json:"password" binding:"gt=8"`
	SamePassword string `json:"samePassword"`
	Role         string `json:"role"`
}

type AccountPatchData struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	SamePassword string `json:"samePassword"`
}

type AccountPatchPromoteData struct {
	Username string `json:"username" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type AccountGetData struct {
	Username string `json:"username" binding:"gt=4"`
	Role     string `json:"role"`
}
