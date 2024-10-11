package repo

import "github.com/jmoiron/sqlx"

// const (
// 	DriverName     = "pgx"
// 	DataSourceName = "host=167.179.66.149 port=5469 user=postgres dbname=yongdb"
// )

func NewDb(driverName string, dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", "host=167.179.66.149 port=5469 user=postgres dbname=yongdb")
	if err != nil {
		panic(err)
	}
	return db, nil
}
