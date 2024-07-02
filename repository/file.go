package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/otaxhu/files-app/models"
)

type FileRepository interface {
	SaveFile(ctx context.Context, file models.File) (err error)
	GetFile(ctx context.Context, id string) (*models.File, error)
	GetFileInfo(ctx context.Context, id string) (*models.File, error)
	DeleteFile(ctx context.Context, id string) error

	InitTables() error
}

type sqlite3FileRepository struct {
	db  *sql.DB
	dir string
}

func NewFileRepository(db *sql.DB, dir string) (FileRepository, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &sqlite3FileRepository{
		db:  db,
		dir: absDir,
	}, nil
}

func (f *sqlite3FileRepository) SaveFile(ctx context.Context, file models.File) (err error) {
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO files (id, file_len, filename) VALUES (?, ?, ?)", file.Id, file.Len, file.Filename)
	if err != nil {
		tx.Rollback()
		return
	}

	newFile, err := os.Create(filepath.Join(f.dir, file.Id))
	if err != nil {
		tx.Rollback()
		return
	}
	defer newFile.Close()

	if _, err = io.Copy(newFile, file.Reader); err != nil {
		tx.Rollback()
		return
	}

	return tx.Commit()
}

func (f *sqlite3FileRepository) GetFile(ctx context.Context, id string) (*models.File, error) {
	row := f.db.QueryRowContext(ctx, "SELECT id, file_len, filename FROM files WHERE id = ?", id)

	retFile := &models.File{}

	err := row.Scan(&retFile.Id, &retFile.Len, &retFile.Filename)
	if err != nil {
		return nil, err
	}

	retFile.Reader, err = os.Open(filepath.Join(f.dir, retFile.Id))
	if err != nil {
		return nil, err
	}

	return retFile, nil
}

func (f *sqlite3FileRepository) GetFileInfo(ctx context.Context, id string) (*models.File, error) {
	row := f.db.QueryRowContext(ctx, "SELECT id, file_len, filename FROM files WHERE id = ?", id)
	retFile := &models.File{}
	return retFile, row.Scan(&retFile.Id, &retFile.Len, &retFile.Filename)
}

func (f *sqlite3FileRepository) DeleteFile(ctx context.Context, id string) error {
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, "DELETE files WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if ra <= 0 {
		tx.Rollback()
		return errors.New("repository: DeleteFile() file not found")
	}

	if err = os.Remove(filepath.Join(f.dir, id)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (f *sqlite3FileRepository) InitTables() (err error) {
	_, err = f.db.Exec(`
CREATE TABLE IF NOT EXISTS files (
	id TEXT(26) PRIMARY KEY
	,file_len INT(20)
	,filename TEXT
);`)
	return
}
