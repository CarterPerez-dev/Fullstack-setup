// AngelaMos | 2026
// database.go

package core

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/carterperez-dev/templates/go-backend/internal/config"
)

type Database struct {
	DB *sqlx.DB
}

func NewDatabase(
	ctx context.Context,
	cfg config.DatabaseConfig,
) (*Database, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(jitteredDuration(cfg.ConnMaxLifetime))
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close() //nolint:errcheck // cleanup on connection failure
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func (d *Database) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := d.DB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func (d *Database) Stats() sql.DBStats {
	return d.DB.Stats()
}

type DBTX interface {
	sqlx.ExtContext
	sqlx.ExecerContext
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(
		ctx context.Context,
		dest any,
		query string,
		args ...any,
	) error
}

func InTx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback() //nolint:errcheck // best-effort rollback on panic
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %w (original: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func InTxWithOptions(
	ctx context.Context,
	db *sqlx.DB,
	opts *sql.TxOptions,
	fn func(tx *sqlx.Tx) error,
) error {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback() //nolint:errcheck // best-effort rollback on panic
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %w (original: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func jitteredDuration(base time.Duration) time.Duration {
	//nolint:gosec // G404: non-security-sensitive jitter for connection pool
	jitter := time.Duration(rand.Int64N(int64(base / 7)))
	return base + jitter
}
