package pkg

import (
	_ "github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"golang.org/x/net/context"
	"html"
	"net"
	_ "os"
)

type DB struct {
	pool *pgxpool.Pool
	ctx  context.Context
}
type ChangeData struct {
	id    int16
	mode  string
	x     int
	y     int
	model string
}

type Message struct {
	Addr string `json:"addr"`
	Text string `json:"text"`
	Id   int    `json:"id"`
}

func (dbp *DB) NewDB() error {
	(*dbp).ctx = context.Background()
	dsn := "postgres://landscape:Ee010800@localhost:5432/landscape"
	var err error
	(*dbp).pool, err = pgxpool.Connect((*dbp).ctx, dsn)
	if err != nil {
		return err
	}

	return nil
}
func (dbp *DB) OnRead(msg string, addr net.Addr) error {
	SQLStatement := "insert into msg(addr,msg) values ($1,$2)"
	conn, err := (*dbp).pool.Acquire((*dbp).ctx)
	defer conn.Release()
	tx, err := conn.Begin(dbp.ctx)
	if err != nil {

		return err
	}
	_, err = tx.Exec((*dbp).ctx, SQLStatement, addr.String(), msg)
	if err != nil {
		defer tx.Rollback((*dbp).ctx)
		return err
	}
	err = tx.Commit(dbp.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (dbp *DB) OnConnection() (interface{}, error) {
	SQLStatement := "select * from msg order by id"
	conn, err := (*dbp).pool.Acquire((*dbp).ctx)
	defer conn.Release()
	if err != nil {

		return []byte(""), err
	}
	tx, err := conn.Begin((*dbp).ctx)
	if err != nil {

		return []byte(""), err
	}
	rows, err := tx.Query((*dbp).ctx, SQLStatement)
	var arr Message
	output := make(map[int]interface{})
	var i = 0
	for rows.Next() {
		err = rows.Scan(&arr.Text, &arr.Addr, &arr.Id)
		if err != nil {
			return []byte(""), err
		}
		arr.Text = html.EscapeString(arr.Text)
		output[i] = arr
		i++
	}
	err = tx.Commit((*dbp).ctx)
	if err != nil {
		return []byte(""), err
	}
	return output, nil
}
