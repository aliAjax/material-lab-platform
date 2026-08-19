package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	inner *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}
	check, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = pool.Ping(check); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Pool{inner: pool}, nil
}

func (p *Pool) Close() {
	p.inner.Close()
}

func (p *Pool) Ready(ctx context.Context) error {
	check, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return p.inner.Ping(check)
}

func (p *Pool) Inner() *pgxpool.Pool {
	return p.inner
}
