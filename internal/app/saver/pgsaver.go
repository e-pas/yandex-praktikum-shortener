package saver

import (
	"context"

	"github.com/jackc/pgx"
)

type pgSaver struct {
	connString string
	connPgx    *pgx.Conn
}

func newPgSaver(conn string) *pgSaver {
	if conn == "" {
		return nil
	}
	return &pgSaver{
		connString: conn,
	}
}

func (pg *pgSaver) Open() error {
	pgxConn, err := pgx.ParseConnectionString(pg.connString)
	if err != nil {
		return err
	}
	pg.connPgx, err = pgx.Connect(pgxConn)
	if err != nil {
		return err
	}
	return nil
}

func (pg *pgSaver) Close() error {
	return pg.connPgx.Close()

}
func (pg *pgSaver) Ping() error {
	err := pg.Open()
	if err != nil {
		return err
	}
	err = pg.connPgx.Ping(context.Background())
	if err != nil {
		return err
	}
	err = pg.Close()
	if err != nil {
		return err
	}
	return nil
}
