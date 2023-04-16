package saver

import (
	"context"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgSaver struct {
	connString string
	pool       *pgxpool.Pool
}

func newPgSaver(conn string) *pgSaver {
	if conn == "" {
		return nil
	}
	return &pgSaver{
		connString: conn,
	}
}

func (pg *pgSaver) createPool() error {
	pgxConnConfig, err := pgxpool.ParseConfig(pg.connString)
	if err != nil {
		return err
	}
	pg.pool, err = pgxpool.NewWithConfig(context.Background(), pgxConnConfig)
	if err != nil {
		return err
	}
	return pg.prepareTable()
}

func (pg *pgSaver) prepareTable() error {
	const createTableSQL = "CREATE TABLE IF NOT EXISTS shrtnr_pair (short VARCHAR(20) PRIMARY KEY, url VARCHAR(80), userid CHAR(32));"
	_, err := pg.pool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) Load(ctx context.Context, data map[string]*config.ShortURL) error {
	const selectSQL = "SELECT short, url, userid FROM shrtnr_pair;"
	if pg.pool == nil {
		if err := pg.createPool(); err != nil {
			return err
		}
	}
	rows, err := pg.pool.Query(ctx, selectSQL)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		shortRec := &config.ShortURL{}
		err := rows.Scan(&shortRec.Short, &shortRec.URL, &shortRec.UserID)
		if err != nil {
			return err
		}
		data[shortRec.Short] = shortRec
	}
	return nil
}

func (pg *pgSaver) Save(ctx context.Context, data *config.ShortURL) error {
	const insertSQL = "INSERT INTO shrtnr_pair (short, url, userid) VALUES ($1, $2, $3);"
	if pg.pool == nil {
		if err := pg.createPool(); err != nil {
			return err
		}
	}
	_, err := pg.pool.Exec(ctx, insertSQL, data.Short, data.URL, data.UserID)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) Ping(ctx context.Context) error {
	if pg.pool == nil {
		if err := pg.createPool(); err != nil {
			return err
		}
	}
	err := pg.pool.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
