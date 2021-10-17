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

	db.Db, err = sql.Open("sqlite3", "bot.db?_key=6HMovdn1osi-7r7")
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
	return db, err
}

func (p *DbMgr) close() {
	p.Db.Close()
}

func (p *DbMgr) get_dynamic_nodes() ([]DynamicNode, error) {
	var ret []DynamicNode
	rows, err := p.Db.Query("SELECT id , ip , port , status FROM dynamic_nodes;")
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
	sql := fmt.Sprintf(`insert into nodes values("%s","%s","%s");`, node.Ip, node.Port, node.Status)
	log.Println(sql)
	_, err := p.Db.Exec(sql)
	if err != nil {
		return false, err
	}
	return true, err
}
func (p *DbMgr) update_dynamic_status(id int64, status string) error {
	sql := fmt.Sprintf(`update node set status="%s" where id=%d;`, status, id)
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
