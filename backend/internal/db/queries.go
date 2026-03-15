package db

func NewQueries(dbClient Client) *Queries {
	return &Queries{
		db: dbClient,
	}
}
