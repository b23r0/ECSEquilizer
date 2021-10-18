package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v2"
)

var g_config GlobalConfig
var g_ecs ECSMgr
var g_db *DbMgr
var g_equilizer EquilizerMgr

type GlobalConfig struct {
	SysConfig struct {
		Listen             string `yaml:"listen"`
		TcpingPort         string `yaml:"tcping_port"`
		WorkInternalSecond int    `yaml:"work_internal_second"`
		HTTPSCallback      string `yaml:"https_callback"`
	} `yaml:"sys_config"`
	AliyunConfig struct {
		AccessKeyID  string `yaml:"access_key_id"`
		AccessSecret string `yaml:"access_secret"`
	} `yaml:"aliyun_config"`
	Authorization []string `yaml:"authorization"`
	StaticNode    []string `yaml:"static_node"`
	EcsTemplate   struct {
		Region                 string `yaml:"region"`
		Imageid                string `yaml:"imageid"`
		Instancetype           string `yaml:"instancetype"`
		Securitygroupid        string `yaml:"securitygroupid"`
		Internetmaxbandwidthin int    `yaml:"internetmaxbandwidthin"`
		Vswitchid              string `yaml:"vswitchid"`
	} `yaml:"ecs_template"`
}

type NodeModel struct {
	ID     string `json:"id"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type CallbackModel struct {
	ID     string `json:"id"`
	IP     string `json:"ip"`
	Action string `json:"action"`
}

type GetNodesRetModel struct {
	Nodes []NodeModel `json:"nodes"`
}

func get_nodes(c echo.Context) (err error) {

	ret := GetNodesRetModel{}
	nodes := g_equilizer.get_nodes()

	for _, s := range nodes {
		var node NodeModel
		node.ID = s.Id
		node.IP = s.Ip
		node.Port = s.Port
		node.Status = s.Status
		node.Type = s.Type

		ret.Nodes = append(ret.Nodes, node)
	}

	return c.JSON(http.StatusOK, &ret)
}

func init_config(filename string) GlobalConfig {

	t := GlobalConfig{}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("load config : " + filename)
	err = yaml.Unmarshal(data, &t)

	if err != nil {
		log.Panicln(err)
	}
	return t
}

func worker() {
	g_equilizer.update_nodes()

	for {
		time.Sleep(time.Duration(g_config.SysConfig.WorkInternalSecond) * time.Second)

		g_equilizer.update_status()

		nodes := g_equilizer.get_nodes()

		normal := 0
		bad := 0
		normal_static := 0
		static := 0

		for _, s := range nodes {

			if s.Type == "static" {
				static++
			}

			if s.Status == "normal" || s.Status == "great" {
				normal++
				if s.Type == "static" {
					normal_static++
				}
			}
			if s.Status == "bad" {
				bad++
			}
		}

		//all static node is normal
		if static == normal_static {

			//uninstall all dynamic

			nodes := g_equilizer.get_nodes()
			for _, s := range nodes {
				if s.Type == "dynamic" {
					g_equilizer.pop_dynamic_node(s.Id)
					g_ecs.delete_ecs(g_config.EcsTemplate.Region, s.InstanceId)
				}
			}

			continue
		}

		// 50% node is bad
		stage := float64(bad) / float64(len(nodes))
		if stage > 0.5 {
			// calc should create dynamic number
			should_num := bad - normal
			// create n dynamic nodes

			for i := 0; i < should_num; i++ {
				ip, instanceid, err_id := g_ecs.create_ecs(g_config.EcsTemplate.Region, g_config.EcsTemplate.Imageid, g_config.EcsTemplate.Instancetype, g_config.EcsTemplate.Securitygroupid, g_config.EcsTemplate.Internetmaxbandwidthin, g_config.EcsTemplate.Vswitchid)
				if err_id != 0 {
					log.Panic("create ecs faild : " + strconv.FormatInt(int64(err_id), 10))
					continue
				}
				g_equilizer.add_dynamic_node(ip, g_config.SysConfig.TcpingPort, "normal", instanceid)
			}
			continue
		}

		// 60% node is normal
		if stage < 0.4 {
			// calc should decrease number
			should_num := normal - bad
			// uninstall n dynamic nodes

			nodes := g_equilizer.get_nodes()
			for _, s := range nodes {
				if s.Type == "dynamic" {
					g_equilizer.pop_dynamic_node(s.Id)
					g_ecs.delete_ecs(g_config.EcsTemplate.Region, s.InstanceId)
				}

				should_num--

				if should_num == 0 {
					break
				}
			}
			continue
		}
	}

}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	g_config = init_config(path.Join(filepath.Dir(os.Args[0]), "config.yaml"))
	g_ecs = ECSMgr{AccessKeyId: g_config.AliyunConfig.AccessKeyID, AccessSecret: g_config.AliyunConfig.AccessSecret}
	g_db, _ = connect_db()

	e := echo.New()
	e.GET("/v1/nodes", get_nodes)
	log.Println("listen to " + g_config.SysConfig.Listen)

	e.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Validator: func(key string, _ echo.Context) (bool, error) {

			for _, s := range g_config.Authorization {
				if s == key {
					return true, nil
				}
			}
			return false, nil
		},
	}))
	go worker()
	e.Logger.Fatal(e.StartTLS(g_config.SysConfig.Listen, path.Join(filepath.Dir(os.Args[0]), "crt/server.crt"), path.Join(filepath.Dir(os.Args[0]), "crt/server.key")))
	g_db.close()
}
