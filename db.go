package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/xeodou/go-sqlcipher"
)

type DbMgr struct {
	Db *sql.DB
}

type DynamicNode struct {
	Id     int
	Ip     string
	Port   string
	Status string
}

//sqlite password 6HMovdn1osi-7r7
func connect_db() (*DbMgr, error) {
	var err error = nil

	var db = new(DbMgr)

	db.Db, err = sql.Open("sqlite3", "a.db?_key=6HMovdn1osi-7r7")
	if err != nil {
		log.Println(err)
		return db, err
	}

	p := "PRAGMA key = '6HMovdn1osi-7r7';"
	_, err = db.Db.Exec(p)
	if err != nil {
		log.Println(err)
		return db, err
	}
	db.init()
	return db, err
}

func (p *DbMgr) init() {
	_, err := p.Db.Exec(`CREATE TABLE "nodes" (
		"id"	INTEGER NOT NULL,
		"ip"	TEXT NOT NULL,
		"port"	TEXT NOT NULL,
		"status"	TEXT NOT NULL,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`)
	if err != nil {
		return
	}

	return
}

func (p *DbMgr) close() {
	p.Db.Close()
}

func (p *DbMgr) get_dynamic_nodes() ([]DynamicNode, error) {
	var ret []DynamicNode
	rows, err := p.Db.Query("SELECT id , ip , port , status FROM nodes;")
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	for rows.Next() {
		var info DynamicNode
		rows.Scan(&info.Id, &info.Ip, &info.Port, &info.Status)
		ret = append(ret, info)
	}

	return ret, err
}

func (p *DbMgr) add_dynamic_node(node DynamicNode) (bool, error) {
	sql := fmt.Sprintf(`insert into nodes(ip , port , status) values("%s","%s","%s");`, node.Ip, node.Port, node.Status)
	log.Println(sql)
	_, err := p.Db.Exec(sql)
	if err != nil {
		return false, err
	}
	return true, err
}
func (p *DbMgr) update_dynamic_status(id int64, status string) error {
	sql := fmt.Sprintf(`update nodes set status="%s" where id=%d;`, status, id)
	_, err := p.Db.Exec(sql)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
func (p *DbMgr) delete_node(id int64) error {
	sql := fmt.Sprintf(`delete from nodes where id=%d;`, id)
	_, err := p.Db.Exec(sql)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
