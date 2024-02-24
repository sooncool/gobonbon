package network

import (
	"context"
	"gobonbon/iface"
	"net"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// (Websocket连接模块, 用于处理 Websocket 连接的读写业务 一个连接对应一个Connection)
type WsConnection struct {
	sync.RWMutex
	wsServer    iface.IServer          //当前Conn属于哪个Server
	conn        *websocket.Conn        //conn 是当前连接的 WebSocket 套接字
	connID      uint64                 // (当前连接的ID 也可以称作为SessionID，ID全局唯一 ，服务端Connection使用,这个是理论支持的进程connID的最大数量) uint64 取值范围：0 ~ 18,446,744,073,709,551,615
	connIdStr   string                 // (字符串的连接id)
	closeFlag   bool                   //当前连接的是否关闭状态
	msgBuffChan chan []byte            // (有缓冲管道，用于读、写两个goroutine之间的消息通信)
	property    map[string]interface{} //(链接属性)
	name        string                 // (链接名称，默认与创建链接的Server/Client的Name一致)
	localAddr   string                 //(当前链接的本地地址)
	remoteAddr  string                 //(当前链接的远程地址)

	onConnStart func(conn iface.IConn) // (当前连接创建时Hook函数)
	onConnStop  func(conn iface.IConn) // (当前连接断开时的Hook函数)
	msgHandler  iface.IMsgHandle       // (消息管理MsgID和对应处理方法的消息管理模块)

	ctx    context.Context    // (告知该链接已经退出/停止的channel)
	cancel context.CancelFunc // (告知该链接已经退出/停止的channel)
}

// (newServerConn :for Server, 创建一个Server服务端特性的连接的方法
// Note: 名字由 NewConnection 更变)
func newWebsocketConn(server iface.IServer, conn *websocket.Conn, connID uint64) iface.IConn {
	// Initialize Conn properties (初始化Conn属性)
	wsConn := &WsConnection{
		wsServer:    server,
		conn:        conn,
		connID:      connID,
		connIdStr:   strconv.FormatUint(connID, 10),
		closeFlag:   false,
		msgBuffChan: nil,
		property:    nil,
		name:        server.ServerName(),
		localAddr:   conn.LocalAddr().String(),
		remoteAddr:  conn.RemoteAddr().String(),
	}

	// lengthField := server.GetLengthField()
	// if lengthField != nil {
	// 	wsConn.frameDecoder = zinterceptor.NewFrameDecoder(*lengthField)
	// }

	// Inherited attributes from server (从server继承过来的属性)
	// wsConn.packet = server.GetPacket()
	// wsConn.onConnStart = server.GetOnConnStart()
	// wsConn.onConnStop = server.GetOnConnStop()
	wsConn.msgHandler = server.GetMsgHandler()

	// Bind the current Connection to the Server's ConnManager (将当前的Connection与Server的ConnManager绑定)
	// wsConn.connManager = server.GetConnMgr()

	// Add the newly created Conn to the connection management (将新创建的Conn添加到链接管理中)
	server.GetConnMgr().Add(wsConn)

	return wsConn
}

func (wsConn *WsConnection) StartReader() {

}

func (wsConn *WsConnection) StartWriter() {

}

// (启动连接，让当前连接开始工作)
func (wsConn *WsConnection) Start() {
	wsConn.ctx, wsConn.cancel = context.WithCancel(context.Background())
	go wsConn.StartReader()
	go wsConn.StartWriter()
	select {
	case <-wsConn.ctx.Done():
		wsConn.finalizer()
		return
	}
}

// (直接将Message数据发送数据给远程的TCP客户端)
func (wsConn *WsConnection) WriteMsg(msgID uint32, data []byte) error {
	return nil
}

func (wsConn *WsConnection) Stop() {
	wsConn.cancel()
}

func (wsConn *WsConnection) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WsConnection) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// 获取当前连接ID
func (wsConn *WsConnection) GetConnID() uint64 {
	return wsConn.connID
}

func (wsConn *WsConnection) GetWsConn() *websocket.Conn {
	return wsConn.conn
}

func (wsConn *WsConnection) GetTCPConnection() net.Conn {
	return nil
}

func (wsConn *WsConnection) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return
	}
	wsConn.closeFlag = true
	wsConn.conn.Close()
	wsConn.cancel() //关闭Writer
	//将链接从连接管理器中删除
	wsConn.wsServer.GetConnMgr().Remove(wsConn)
	//关闭该链接全部管道
	close(wsConn.msgBuffChan)
}
