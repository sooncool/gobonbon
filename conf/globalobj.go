package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ServerModeTcp       = "tcp"
	ServerModeWebsocket = "websocket"
	ServerModeUdp       = "udp"
)

/*
存储一切有关gobonbon框架的全局参数，供其他模块使用
一些参数也可以通过 用户根据 gobonbon.json来配置
*/
type GlobalObj struct {
	Host    string //当前服务器主机IP
	TcpPort int    //当前服务器主机监听端口号
	WsPort  int    //当前服务器主机websocket监听端口
	Name    string //当前服务器名称
	Version string //当前gobonbon版本号

	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	MaxMsgChanLen    int
	WorkerPoolSize   uint64 //业务工作Worker池的数量
	MaxWorkerTaskLen uint64 //业务工作Worker对应负责的任务队列最大任务存储数量

	Mode string
}

/*
定义一个全局的对象
*/
var GlobalObject *GlobalObj

// PathExists Check if a file exists.(判断一个文件是否存在)
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetConfigFilePath() string {
	configFilePath := os.Getenv("GOBONBON_CONFIG_FILE_PATH")
	if configFilePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configFilePath = filepath.Join(pwd, "/conf/gobon.json")
	}
	var err error
	configFilePath, err = filepath.Abs(configFilePath)
	if err != nil {
		panic(err)
	}
	return configFilePath

}

// 读取用户的配置文件
func (g *GlobalObj) Reload() {
	confFilePath := GetConfigFilePath()
	if confFileExists, _ := PathExists(confFilePath); !confFileExists {
		return
	}

	data, err := os.ReadFile(confFilePath)
	fmt.Println(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	//将json数据解析到struct中
	//fmt.Printf("json :%s\n", data)
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

/*
提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:    "gobonbon",
		Version: "V1.0",
		TcpPort: 7777,
		WsPort:  9000,
		Host:    "0.0.0.0",

		MaxPacketSize:    4096,
		MaxConn:          12000,
		MaxMsgChanLen:    1024,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
	}

	//从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
