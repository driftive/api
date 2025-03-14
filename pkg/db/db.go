package db

import (
	"context"
	"driftive.cloud/api/pkg/config"
	"driftive.cloud/api/pkg/repository/queries"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"strconv"
)

type TxType string

type DB struct {
	Config     *pgxpool.Config
	Pool       *pgxpool.Pool
	rawQueries *queries.Queries
}

func NewDB(cfg config.Config) *DB {
	hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s/%s?",
		cfg.Database.User, cfg.Database.Password, hostPort, cfg.Database.Database)
	dbConfig, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Panic("Failed to parse database config")
	}

	dbConfig.MaxConns = cfg.Database.Connections
	dbConfig.MinConns = cfg.Database.Connections

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		log.Panic("Failed to connect to database")
	}

	return &DB{
		Config:     dbConfig,
		Pool:       pool,
		rawQueries: queries.New(pool),
	}
}

func (d *DB) Queries(ctx context.Context) *queries.Queries {
	if ctx.Value("tx") != nil {
		return queries.New(d.Pool).WithTx(ctx.Value("tx").(pgx.Tx))
	}
	return d.rawQueries
}

func (d *DB) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if ctx.Value(TxType("tx")) != nil {
		return errors.New("transaction already in progress")
	}
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	newCtx := context.WithValue(ctx, TxType("tx"), tx)
	defer tx.Rollback(ctx)

	if err := fn(newCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
