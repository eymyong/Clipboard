package dbclipboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/jmoiron/sqlx"
)

type RepoDbClipboard struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) repo.RepositoryClipboard {
	return &RepoDbClipboard{db: db}
}

func (d *RepoDbClipboard) Create(ctx context.Context, clip model.Clipboard) error {
	tx, err := repo.Begin(ctx, d.db)
	if err != nil {
		return err
	}

	rows, err := tx.Exec("insert into CLIPBOARDS (id,user_id,text) values ($1,$2,$3)", clip.Id, clip.UserId, clip.Text)
	if err != nil {
		return fmt.Errorf("error inserting into clipboards: %w", err)
	}

	c, err := rows.RowsAffected()
	if err != nil {
		return repo.Rollback(tx, err)
	}
	if c != 1 {
		return repo.Rollback(tx, repo.UnexpectedRows(c))
	}

	return tx.Commit()
}

func (d *RepoDbClipboard) GetAll(ctx context.Context) ([]model.Clipboard, error) {
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

func (d *RepoDbClipboard) GetAllUserClipboards(ctx context.Context, userId string) ([]model.Clipboard, error) {
	rows, err := d.db.QueryContext(ctx, "select * from CLIPBOARDS where user_id = ($1)", userId)
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

func (d *RepoDbClipboard) GetById(ctx context.Context, id string) (model.Clipboard, error) {
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

func (d *RepoDbClipboard) GetUserClipboard(ctx context.Context, id, userId string) (model.Clipboard, error) {
	rows, err := d.db.QueryContext(ctx, "select text from CLIPBOARDS where id = ($1) and user_id = ($2)", id, userId)
	if err != nil {
		return model.Clipboard{}, err
	}

	var text string
	c := 0
	for rows.Next() {
		err := rows.Scan(&text)
		if err != nil {
			return model.Clipboard{}, err
		}

		c++
	}

	if c != 0 {
		return model.Clipboard{}, repo.UnexpectedRows(c)
	}

	return model.Clipboard{
		Id:     id,
		UserId: userId,
		Text:   text,
	}, nil
}

func (d *RepoDbClipboard) Update(ctx context.Context, id, newText string) error {
	tx, err := repo.Begin(ctx, d.db)
	if err != nil {
		return err
	}

	result, err := tx.Exec("update CLIPBOARDS c set text = ($1) where c.id = ($2)", newText, id)
	if err != nil {
		return fmt.Errorf("query update err: %w", err)
	}

	c, err := result.RowsAffected()
	if err != nil {
		return repo.Rollback(tx, err)
	}

	if c != 1 {
		return repo.Rollback(tx, repo.UnexpectedRows(c))
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit 'tx': %w", err)
	}

	return nil
}

func (d *RepoDbClipboard) UpdateUserClipboard(ctx context.Context, id, userId, newText string) error {
	return nil
}

func (d *RepoDbClipboard) Delete(ctx context.Context, id string) error {
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
		return fmt.Errorf("rowsaffected delete clipboards err: %w", err)
	}

	if c != 1 {
		return fmt.Errorf("unexpected row: %d", c)

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit 'tx': %w", err)
	}

	return nil
}

func (d *RepoDbClipboard) DeleteUserClipboard(ctx context.Context, id, userId string) error {
	tx, err := repo.Begin(ctx, d.db)
	if err != nil {
		return err
	}

	_, err = tx.Exec("delete from USERS where id = ($1) and user_id = ($2)", id, userId)
	return err
}

func (d *RepoDbClipboard) DeleteAll(ctx context.Context) error {
	return errors.New("not implemented")
}
