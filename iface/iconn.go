package iface

import (
	"net"

	"github.com/gorilla/websocket"
)

type IConn interface {
	Start()               //启动连接，让当前连接开始工作
	Stop()                //停止连接，结束当前连接状态
	LocalAddr() net.Addr  //获取本地客户端的TCP状态 IP Port'
	RemoteAddr() net.Addr //获取远程客户端的TCP状态 IP Port'
	WriteMsg(msgId uint32, args []byte) error
	GetConnID() uint64 //获取当前连接ID

	GetTCPConnection() net.Conn
	GetWsConn() *websocket.Conn // 从当前连接中获取原始的websocket连接)
}
