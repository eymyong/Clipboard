package dbclipboard

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/jmoiron/sqlx"
)

type RepoDB struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) repo.RepositoryClipboard {
	return &RepoDB{db: db}
}

func (d *RepoDB) Create(ctx context.Context, clip model.Clipboard) error {
	_, err := d.db.QueryContext(ctx, "insert into CLIPBOARDS (id,user_id,text) values ($1,$2,$3)", clip.Id, clip.UserId, clip.Text)
	if err != nil {
		return fmt.Errorf("query insert err: %w", err)
	}
	return nil
}

func (d *RepoDB) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	rows, err := d.db.QueryContext(ctx, "select * from CLIPBOARDS")
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("query select err: %w", err)
	}

	var clipboard []model.Clipboard
	for rows.Next() {
		var clip model.Clipboard
		err := rows.Scan(&clip.Id, &clip.UserId, &clip.Text)
		if err != nil {
			return []model.Clipboard{}, fmt.Errorf("scan err: %w", err)
		}

		clipboard = append(clipboard, clip)
	}

	return clipboard, nil
}

func (d *RepoDB) GetById(ctx context.Context, id string) (model.Clipboard, error) {
	rows, err := d.db.QueryContext(ctx, "select * from CLIPBOARDS c where c.id = ($1)", id)
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("query select err: %w", err)
	}
	// var clipID int = -1
	var clipboard model.Clipboard
	for rows.Next() {
		err := rows.Scan(&clipboard.Id, &clipboard.UserId, &clipboard.Text)
		if err != nil {
			return model.Clipboard{}, fmt.Errorf("scan err: %w", err)
		}
	}

	if clipboard.Id != id {
		return model.Clipboard{}, fmt.Errorf("invalid id: %s", id)
	}

	return clipboard, nil
}

func (d *RepoDB) Update(ctx context.Context, id string, newText string) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, "update CLIPBOARDS c set text = ($1) where c.id = ($2)", newText, id)
	if err != nil {
		return fmt.Errorf("query update err: %w", err)
	}

	c, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsaffected err: %w", err)
	}

	if c != 1 {
		err := tx.Rollback()
		if err != nil {
			return fmt.Errorf("failed to rollback:%w", err)
		}

		return fmt.Errorf("unexpected rows: %d", c)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commid 'tx': %w", err)
	}

	return nil
}

func (d *RepoDB) Delete(ctx context.Context, id string) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "delete from CLIPBOARDS where id = ($1)", id)
	if err != nil {
		return fmt.Errorf("query delete err: %w", err)
	}

	c, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rowsaffected err: %w", err)
	}

	if c != 1 {
		return fmt.Errorf("unexpected row: %d", c)

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commid 'tx': %w", err)
	}

	return nil
}

func (d *RepoDB) DeleteAll(ctx context.Context) error {

	return nil
}
