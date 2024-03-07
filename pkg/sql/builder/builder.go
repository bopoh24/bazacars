package builder

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
)

type Postgres struct {
	db      *sql.DB
	builder *sq.StatementBuilderType
}

// NewPostgresBuilder is a wrapper around squirrel.StatementBuilderType
func NewPostgresBuilder(psqlConn string) (*Postgres, error) {
	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	dbCache := sq.NewStmtCache(db)
	if err != nil {
		return nil, err
	}
	builder := sq.StatementBuilder.RunWith(dbCache).PlaceholderFormat(sq.Dollar)
	return &Postgres{db: db, builder: &builder}, nil
}

// Builder returns a new statement builder
func (p *Postgres) Builder() *sq.StatementBuilderType {
	return p.builder
}

// Tx returns a new transaction
func (p *Postgres) Tx(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// Close closes the database connection
func (p *Postgres) Close() error {
	return p.db.Close()
}
