package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func NewDb(driverName string, dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", "host=167.179.66.149 port=5469 user=postgres dbname=yongdb")
	if err != nil {
		panic(err)
	}

	return db, nil
}

func Begin(
	ctx context.Context,
	database interface {
		BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
	},
) (
	*sql.Tx,
	error,
) {
	return database.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
}

// Rollback rolls back transaction if err is not nil.
// If err is nil, tx is not rolled back.
// If err is not nil and Rollback is ok, the original err is returned.
// If err is not nil and Rollback has error, a wrapped error is returned.
func Rollback(tx interface{ Rollback() error }, err error) error {
	if err == nil {
		return nil
	}

	errRollback := tx.Rollback()
	if errRollback != nil {
		return fmt.Errorf("failed to rollback after error '%w': %w", err, errRollback)
	}

	return err
}

func UnexpectedRows[T int | int64](c T) error {
	return fmt.Errorf("unexpected row count: %d", c)
}
