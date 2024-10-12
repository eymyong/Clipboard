package dbuser

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

type RepoDbUser struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) repo.RepositoryUser {
	return &RepoDbUser{db: db}
}

func (d *RepoDbUser) Create(ctx context.Context, user model.User) (model.User, error) {
	tx, err := repo.Begin(ctx, d.db)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to begin tx: %w", err)
	}

	rows, err := tx.Query("insert into USERS (id,username,password) values ($1,$2,$3) returning username, password", user.Id, user.Username, user.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("query insert `users` err: %w", err)
	}

	var username, password string
	c := 0
	for rows.Next() {
		err = rows.Scan(&username, &password)
		if err != nil {
			return model.User{}, repo.Rollback(tx, err)
		}

		c++
	}

	if c != 1 {
		return model.User{}, repo.Rollback(tx, repo.UnexpectedRows(c))
	}

	err = tx.Commit()
	if err != nil {
		return model.User{}, fmt.Errorf("failed to commit new user: %w", err)
	}

	return model.User{
		Username: username,
		Password: password,
	}, nil
}

func (d *RepoDbUser) GetPassword(ctx context.Context, username string) ([]byte, error) {
	rows, err := d.db.QueryContext(ctx, "select password from USERS where username = ($1)", username)
	if err != nil {
		return nil, fmt.Errorf("query select err: %w", err)
	}

	var user model.User
	for rows.Next() {
		err := rows.Scan(&user.Password)
		if err != nil {
			return nil, fmt.Errorf("scan err: %w", err)
		}
	}

	return []byte(user.Password), nil
}

func (d *RepoDbUser) GetById(ctx context.Context, id string) (model.User, error) {
	rows, err := d.db.QueryContext(ctx, "select * from USERS where id = ($1)", id)
	if err != nil {
		return model.User{}, fmt.Errorf("query select err: %w", err)
	}

	var user model.User
	c := 0
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Username, &user.Password)
		if err != nil {
			return model.User{}, fmt.Errorf("scan err: %w", err)
		}

		c++
	}

	if c != 1 {
		return model.User{}, repo.UnexpectedRows(c)
	}

	if user.Id != id {
		return model.User{}, fmt.Errorf("invalid id: %s", id)
	}

	return user, nil
}

// postgres ไม่ต้องใช้ GetUserId ?
func (d *RepoDbUser) GetUserId(ctx context.Context, username string) (string, error) {
	rows, err := d.db.QueryContext(ctx, "select id from USERS where username = ($1)", username)
	if err != nil {
		return "", fmt.Errorf("query select err: %w", err)
	}

	var userId string
	c := 0
	for rows.Next() {
		err := rows.Scan(&userId)
		if err != nil {
			return "", fmt.Errorf("scan err: %w", err)
		}

		c++
	}

	if c != 1 {
		return "", repo.UnexpectedRows(c)
	}

	return userId, nil
}

func (d *RepoDbUser) UpdateUsername(ctx context.Context, id string, newUsername string) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	result, err := tx.Exec("update USERS u set username = ($1) where u.id = ($2)", newUsername, id)
	if err != nil {
		return fmt.Errorf("query update username err: %w", err)
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

func (d *RepoDbUser) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, "update USERS u set password = ($1) where u.id = ($2)", newPassword, id)
	if err != nil {
		return fmt.Errorf("query update password err: %w", err)
	}

	c, err := result.RowsAffected()
	if err != nil {
		return repo.Rollback(tx, err)
	}
	if c != 1 {
		return repo.Rollback(tx, fmt.Errorf("unexpected rows affected: %d", c))
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit 'tx': %w", err)
	}

	return nil
}

func (d *RepoDbUser) Delete(ctx context.Context, id string) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, "delete from USERS where id = ($1)", id)
	if err != nil {
		return fmt.Errorf("query delete users err: %w", err)
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
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}
