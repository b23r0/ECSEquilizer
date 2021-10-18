package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type NodeRecord struct {
	Id         string
	Ip         string
	Port       string
	Type       string
	Status     string
	InstanceId string
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
}

func (p *EquilizerMgr) add_dynamic_node(ip string, port string, status string, instanceid string) string {

	var node DynamicNode
	node.Ip = ip
	node.Port = port
	node.Status = status

	id, err := g_db.add_dynamic_node(&node)

	if err != nil || id == -1 {
		log.Panic(err)
	}

	var record NodeRecord

	record.Id = "D" + strconv.FormatInt(id, 10)
	record.Ip = ip
	record.Port = port
	record.Status = status
	record.Type = "dynamic"

	p.NodeList = append(p.NodeList, record)
	return record.Id
}

func (p *EquilizerMgr) pop_dynamic_node(id string) NodeRecord {
	var ret NodeRecord
	for i := 0; i < len(p.NodeList); i++ {
		if p.NodeList[i].Id == id {
			ret = p.NodeList[i]
			p.NodeList = append(p.NodeList[:i], p.NodeList[i+1:]...)
			break
		}
	}

	sid := strings.Split(ret.Id, "D")
	iid, _ := strconv.Atoi(sid[1])
	g_db.delete_node(int64(iid))

	return ret

}

func (p *EquilizerMgr) update_node_status(id string, status string) {
	for i := 0; i < len(p.NodeList); i++ {
		if p.NodeList[i].Id == id {
			p.NodeList[i].Status = status

			if p.NodeList[i].Type == "dynamic" {
				sid := strings.Split(p.NodeList[i].Id, "D")
				iid, _ := strconv.Atoi(sid[1])
				g_db.update_dynamic_status(int64(iid), status)
			}
		}
	}
}

func (p *EquilizerMgr) get_nodes() []NodeRecord {
	return p.NodeList
}

func (p *EquilizerMgr) ping(ip string, port string) int64 {
	pinger := NewTCPing()

	iport, _ := strconv.Atoi(port)

	target := Target{
		Timeout:  5 * time.Second,
		Interval: 1 * time.Second,
		Host:     ip,
		Port:     iport,
		Counter:  4,
		Proxy:    "",
		Protocol: TCP,
	}

	pinger.SetTarget(&target)
	pingdone := pinger.Start()

	var a chan os.Signal
	select {
	case <-pingdone:
		break
	case <-a:
		break
	}
	result := pinger.Result()

	if result.SuccessCounter == 0 {
		return -1
	}

	avg := result.TotalDuration / time.Duration(result.SuccessCounter)

	return avg.Milliseconds()
}

func (p *EquilizerMgr) update_status() {

	nodes := p.get_nodes()

	for _, s := range nodes {
		t := p.ping(s.Ip, s.Port)
		if t == -1 {
			p.update_node_status(s.Id, "Bad")
			continue
		} else {
			if t < 100 {
				p.update_node_status(s.Id, "great")
			} else if t < 2000 {
				p.update_node_status(s.Id, "normal")
			} else {
				p.update_node_status(s.Id, "bad")
			}
		}
	}

}
