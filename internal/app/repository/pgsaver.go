package repository

import (
	"context"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgSaver struct {
	connString string
	pool       *pgxpool.Pool
}

const createTableSQL = "CREATE TABLE IF NOT EXISTS shrtnr_pair (short VARCHAR(20) PRIMARY KEY, url VARCHAR(80), userid CHAR(32), deleted BOOLEAN);"
const selectSQL = "SELECT short, url, userid, deleted FROM shrtnr_pair;"
const insertSQL = "INSERT INTO shrtnr_pair (short, url, userid, deleted) VALUES ($1, $2, $3, $4);"
const updateSQL = "UPDATE shrtnr_pair SET url = $2, userid = $3, deleted = $4 WHERE short = $1;"

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
	_, err := pg.pool.Exec(context.Background(), createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) Load(ctx context.Context, data map[string]*model.ShortURL) error {
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
		shortRec := &model.ShortURL{}
		err := rows.Scan(&shortRec.Short, &shortRec.URL, &shortRec.UserID, &shortRec.Deleted)
		if err != nil {
			return err
		}
		data[shortRec.Short] = shortRec
	}
	return nil
}

func (pg *pgSaver) Save(ctx context.Context, data model.ShortURL) error {
	_, err := pg.pool.Exec(ctx, insertSQL, data.Short, data.URL, data.UserID, data.Deleted)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) UpdateBatch(ctx context.Context, data []*model.ShortURL) error {
	return pg.batchUpsert(ctx, updateSQL, data)
}

func (pg *pgSaver) SaveBatch(ctx context.Context, data []*model.ShortURL) error {
	return pg.batchUpsert(ctx, insertSQL, data)
}

func (pg *pgSaver) Ping(ctx context.Context) error {
	pctx := ctx
	if ctx == nil {
		pctx = context.Background()
	}
	err := pg.pool.Ping(pctx)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) batchUpsert(ctx context.Context, sqlStatement string, data []*model.ShortURL) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	btch := &pgx.Batch{}

	for _, rec := range data {
		btch.Queue(sqlStatement, rec.Short, rec.URL, rec.UserID, rec.Deleted)
	}
	bres := tx.SendBatch(ctx, btch)

	for range data {
		_, qerr := bres.Exec()
		if qerr != nil {
			return qerr
		}
	}
	err = bres.Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
