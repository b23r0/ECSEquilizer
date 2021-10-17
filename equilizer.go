package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type NodeRecord struct {
	Id     string
	Ip     string
	Port   string
	Type   string
	Status string
}

type EquilizerMgr struct {
	NodeList []NodeRecord
}

func (p *EquilizerMgr) update_nodes() {

	p.NodeList = p.NodeList[0:0]

	i := 0
	for _, s := range g_config.StaticNode {
		var record NodeRecord
		tmp := strings.Split(s, ":")
		record.Ip = tmp[0]
		record.Port = tmp[1]
		record.Id = "S" + fmt.Sprint(i)
		record.Type = "static"
		i++
		p.NodeList = append(p.NodeList, record)
	}

	dnodes, err := g_db.get_dynamic_nodes()

	if err != nil {
		log.Panic(err)
	}

	for _, s := range dnodes {
		var record NodeRecord
		record.Id = "D" + fmt.Sprint(s.Id)
		record.Ip = s.Ip
		record.Port = s.Port
		record.Status = s.Status
		record.Type = "dynamic"
		p.NodeList = append(p.NodeList, record)
	}

	p.update_status()
}

func (p *EquilizerMgr) get_nodes() []NodeRecord {
	return p.NodeList
}

func (p *EquilizerMgr) update_status() {
	pinger := NewTCPing()

	target := Target{
		Timeout:  5 * time.Second,
		Interval: 1 * time.Second,
		Host:     "127.0.0.1",
		Port:     1234,
		Counter:  4,
		Proxy:    "",
		Protocol: TCP,
	}

	pinger.SetTarget(&target)
}
