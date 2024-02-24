package network

import (
	"fmt"
	"gobonbon/conf"
	"gobonbon/iface"
	"gobonbon/msgparser"
	"gobonbon/router"
	"gobonbon/util"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Server struct {
	Name      string // Name of the server (服务器的名称)
	IPVersion string //tcp4 or other
	IP        string // IP version (e.g. "tcp4") - 服务绑定的IP地址
	Port      int    // IP address the server is bound to (服务绑定的端口)
	WsPort    int    // 服务绑定的websocket 端口 (Websocket port the server is bound to)

	// (异步捕获链接关闭状态)
	// exitChan chan struct{}

	msgHandler iface.IMsgHandle //当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	msgParser  iface.IMsgParser
	ConnMgr    iface.IConnManager //当前Server的链接管理器

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool

	cID uint64 // 连接ID

	// websocket
	upgrader *websocket.Upgrader
	// websocket connection authentication
	websocketAuth func(r *http.Request) error
}

func NewServerWithConfig() *Server {
	// conf.GlobalObject.Reload()
	s := &Server{
		Name:       conf.GlobalObject.Name,
		IPVersion:  "tcp",
		IP:         conf.GlobalObject.Host,
		Port:       conf.GlobalObject.TcpPort,
		WsPort:     conf.GlobalObject.WsPort,
		msgHandler: router.NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

// (开启网络服务)
func (s *Server) Start() {
	// (启动worker工作池机制)
	s.msgHandler.StartWorkerPool()
	// (开启一个go去做服务端Listener业务)
	switch conf.GlobalObject.Mode {
	case conf.ServerModeTcp:
		fmt.Printf(" server , name %d", 11)
		go s.ListenTcpConn()
	case conf.ServerModeWebsocket:
		go s.ListenWebsocketConn()
	default:
		go s.ListenTcpConn()
	}
}

// Stop stops the server (停止服务)
func (s *Server) Stop() {
	fmt.Printf("[STOP] Zinx server , name %s", s.Name)
	// (将其他需要清理的连接信息或者其他信息 也要一并停止或者清理)
	s.ConnMgr.ClearConn()
}

func (s *Server) ListenTcpConn() {
	fmt.Printf(" server , name %d", 11)
	// 1. Get a TCP address
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Printf("[START] resolve tcp addr err: %v\n", err)
		return
	}
	listener, err := net.ListenTCP(s.IPVersion, addr)
	if err != nil {
		fmt.Printf("[ListenTCP is err: %v\n", err)
		panic(err)
	}

	msgParser := msgparser.NewMsgParser()

	s.msgParser = msgParser

	go func() {
		for {
			//设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= conf.GlobalObject.MaxConn {
				continue
			}
			// (阻塞等待客户端建立连接请求)
			conn, err := listener.AcceptTCP()
			if err != nil {
				continue
			}

			newCid := atomic.AddUint64(&s.cID, 1)
			fmt.Println("555")
			dealConn := newTcpConn(s, conn, newCid, s.msgParser, s.msgHandler)
			fmt.Println("555")
			go s.StartConn(dealConn)
		}
	}()

	select {}
}

func (s *Server) ListenWebsocketConn() {
	fmt.Printf(" server , name %d", 11)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// (设置服务器最大连接控制,如果超过最大连接，则等待)
		fmt.Printf(" server , name %d", 111)
		if s.ConnMgr.Len() >= conf.GlobalObject.MaxConn {
			fmt.Printf("Exceeded the maxConnNum:%d, Wait:%d", conf.GlobalObject.MaxConn, util.AcceptDelay)
			util.AcceptDelay.Delay()
			return
		}
		fmt.Printf(" server , name %d", 1111)
		// (如果需要 websocket 认证请设置认证信息)
		if s.websocketAuth != nil {
			err := s.websocketAuth(r)
			if err != nil {
				fmt.Printf(" websocket auth err:%v", err)
				w.WriteHeader(401)
				util.AcceptDelay.Delay()
				return
			}
		}

		// (升级成 websocket 连接)
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(500)
			util.AcceptDelay.Delay()
			return
		}
		fmt.Printf(" server , name %d", 22)
		util.AcceptDelay.Reset()
		// 5. 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
		newCid := atomic.AddUint64(&s.cID, 1)
		wsConn := newWebsocketConn(s, conn, newCid)
		fmt.Printf(" server , name %d", newCid)
		go s.StartConn(wsConn)
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.IP, s.WsPort), nil)
	if err != nil {
		panic(err)
	}
}

func (s *Server) StartConn(conn iface.IConn) {
	// 开始处理当前连接的业务
	conn.Start()
}

// 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server) AddRouter(msgId uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
	fmt.Println("Add Router succ! ")
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

func (s *Server) ServerName() string {
	return s.Name
}

func (s *Server) GetMsgHandler() iface.IMsgHandle {
	return s.msgHandler
}
