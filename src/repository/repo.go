package repository

type Repo struct {
	DB *Database
}

func NewRepo(db *Database) *Repo {
	return &Repo{DB: db}
}
