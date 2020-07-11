package pkg

import (
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx"
	"github.com/jackc/pgx/pgxpool"
	"golang.org/x/net/context"
	"net"
	_ "os"
)

type DB struct {
	pool *pgxpool.Pool
	ctx context.Context
}
type ChangeData struct{
	id int16
	mode string
	x int
	y int
	model string
}

func (dbp *DB) NewDB() error{
	(*dbp).ctx =context.Background()
	dsn:="postgres://landscape:Ee010800@localhost:5432/landscape"
	var err error
	(*dbp).pool, err = pgxpool.Connect((*dbp).ctx,dsn)
	if err!=nil {
		return err
	}
	return nil
}
func (dbp *DB) OnRead(msg string,addr net.Addr) error{
	conn,err := (*dbp).pool.Acquire((*dbp).ctx)
	defer conn.Release()
	tx, err := conn.Begin(dbp.ctx)
	if err != nil {

		return err
	}
	_, err =tx.Exec((*dbp).ctx, "insert into msg(addr,msg) values ($1,$2)",addr.String(),msg)
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

func (dbp *DB) OnConnection() (string,error) {
	conn,err := (*dbp).pool.Acquire((*dbp).ctx)
	defer conn.Release()
	if err != nil {

		return "",err
	}
	tx, err := conn.Begin((*dbp).ctx)
	if err != nil {

		return "",err
	}
	rows, err :=tx.Query((*dbp).ctx, "select * from msg")
	var text string
	var addr string
	var id int64
	var output = ""
	for rows.Next() {
		err = rows.Scan(&text,&addr,&id)
		if err != nil {
			return "",err
		}
		fmt.Print("row")
		output += "<p>"+addr+": "+text+"</p>"

	}
	err = tx.Commit((*dbp).ctx)
	if err != nil {
		return "",err
	}
	return output,nil
}
func (dbp *DB) OnReadJSON(msg []byte,addr net.Addr) error{
	conn,err := (*dbp).pool.Acquire((*dbp).ctx)
	if err != nil {
		return err
	}
	var arr ChangeData
	err = json.Unmarshal(msg, &arr)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(dbp.ctx)
	if err != nil {

		return err
	}
	_, err =tx.Exec((*dbp).ctx, "insert into " +
		"player_changes(mode,x,y,model,addr) " +
		"values ($1,$2,$3,$4,$5)",arr.mode,arr.x,arr.y,arr.model,addr.String(),msg)
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

func (dbp *DB) OnJSONConnection() (  []byte,error) {
	conn,err := (*dbp).pool.Acquire((*dbp).ctx)
	defer conn.Release()
	if err != nil {
		return nil,err
	}
	tx, err := conn.Begin((*dbp).ctx)

	if err != nil {

		return nil,err
	}

	rows, err :=tx.Query((*dbp).ctx, "select * from msg")

	var row ChangeData
	arr:= make(map[int]ChangeData)
	var i int = 0

	for rows.Next() {
		err = rows.Scan(&row)
		if err != nil {
			return nil,err
		}
		arr[i] = row
		fmt.Print("row")
		i++
	}
	output,err:=json.Marshal(arr)
	err = tx.Commit((*dbp).ctx)
	if err != nil {
		return nil,err
	}
	return output,nil
}
