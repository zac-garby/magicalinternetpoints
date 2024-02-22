package common

type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash []byte
}

type Site struct {
	ID               int
	Title            string
	URL              string
	ScoreDescription string
	Sources          []*PointSource
}

type Account struct {
	Site       *Site
	Username   string
	ProfileURL string
}

type PointSource struct {
	ID          int
	Name        string
	Description string

	LowUpper    int
	MediumUpper int
	LowRate     float64
	MediumRate  float64
	HighRate    float64

	Site *Site
}

type Points struct {
	Source      *PointSource
	LastUpdated uint64
	Raw         int
	Real        int
}

type AccountPoints struct {
	Account *Account
	Points  []*Points
	Total   int
}
