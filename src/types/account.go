package types

type AccountPostData struct {
	Username     string `json:"Username" binding:"gt=4"`
	Password     string `json:"Password" binding:"gt=8"`
	SamePassword string `json:"SamePassword"`
	Role         string `json:"Role"`
}

type AccountPatchData struct {
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	SamePassword string `json:"SamePassword"`
}

type AccountPatchPromoteData struct {
	Username string `json:"Username" binding:"required"`
	Role     string `json:"Role" binding:"required"`
}

type AccountGetData struct {
	Username string `json:"Username" binding:"gt=4"`
	Role     string `json:"Role"`
}
