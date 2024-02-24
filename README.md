一、gobonbon配置文件
{
		Name:    "gobonbon",
		Version: "V1.0",
		TcpPort: 7777,
		Host:    "0.0.0.0",

		MaxPacketSize:    4096,
		MaxConn:          12000,
		MaxMsgChanLen:    1024,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
}
Name:    服务器应用名称
Version: 版本号
TcpPort: 服务器监听端口
Host:    服务器IP

MaxPacketSize:    4096,
MaxConn:          允许的客户端链接最大数量
MaxMsgChanLen:    消息最大长度
WorkerPoolSize:   工作任务池最大工作Goroutine数量
MaxWorkerTaskLen: 

二、框架结构
1、conf 		配置文件、框架的全局参数
2、Demo 		测试服务器运行
3、iface  		连接方法接口和框架其他方法的接口
4、log 			简单的log封装
5、msgparser 	TCP消息封装和拆包，防止TCP粘包   
6、network 		TCP、WS、UDP连接封装
7、recordfile
8、router 		路由方法封装
