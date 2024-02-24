package network

//golang标准库的网络模块足够强大易用了，我们只做
import (
	"context"
	"errors"
	"fmt"
	"gobonbon/conf"
	"gobonbon/iface"
	"gobonbon/msgparser"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ConnSet map[net.Conn]struct{}

type TCPConn struct {
	sync.RWMutex
	TCPServer  iface.IServer //当前Conn属于哪个Server
	conn       net.Conn      //当前连接socket Tcp套接字
	connID     uint64        //当前连接的ID 也可以称作为SessionID，ID全局唯一
	closeFlag  bool          //当前连接的状态
	writeChan  chan []byte   // (有缓冲管道，用于读、写两个goroutine之间的消息通信)
	MsgHandler iface.IMsgHandle
	msgParser  iface.IMsgParser
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
}

// 初始化链接模块的方法
func newTcpConn(server iface.IServer, conn *net.TCPConn, connID uint64, msgParser iface.IMsgParser, msgHandler iface.IMsgHandle) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.TCPServer = server
	tcpConn.conn = conn
	tcpConn.connID = connID
	tcpConn.writeChan = make(chan []byte, conf.GlobalObject.MaxMsgChanLen)
	tcpConn.closeFlag = false
	tcpConn.msgParser = msgParser
	tcpConn.MsgHandler = msgHandler
	//将新创建的Conn添加到链接管理中
	tcpConn.TCPServer.GetConnMgr().Add(tcpConn)
	return tcpConn
}

func (tcpConn *TCPConn) StartReader() {
	defer fmt.Printf("%s [conn Writer exit!]\n", tcpConn.RemoteAddr().String())
	defer tcpConn.Stop()
	for {
		select {
		case <-tcpConn.ctx.Done():
			return
		default:
			// 创建拆包解包的对象
			hlen := tcpConn.msgParser.GetHeadLen()

			//读取客户端的Msg head
			headData := make([]byte, hlen)
			if _, err := io.ReadFull(tcpConn.conn, headData); err != nil {
				fmt.Println("read msg head error ", err)
				return
			}
			//fmt.Printf("read headData %+v\n", headData)

			//拆包，得到msgid 和 datalen 放在msg中
			msg, err := tcpConn.msgParser.Decode(headData)
			if err != nil {
				fmt.Println("unpack error ", err)
				return
			}

			//根据 dataLen 读取 data，放在msg.Data中
			var data []byte
			if msg.GetDataLen() > 0 {
				data = make([]byte, msg.GetDataLen())
				if _, err := io.ReadFull(tcpConn.conn, data); err != nil {
					fmt.Println("read msg data error ", err)
					return
				}
			}
			msg.SetData(data)
			fmt.Println("555")
			//得到当前客户端请求的Request数据
			req := NewRequest(tcpConn, msg)

			if conf.GlobalObject.WorkerPoolSize > 0 {
				//已经启动工作池机制，将消息交给Worker处理
				tcpConn.MsgHandler.SendMsgToTaskQueue(req)
			} else {
				//从绑定好的消息和对应的处理方法中执行对应的Handle方法
				go tcpConn.MsgHandler.DoMsgHandler(req)
			}
		}
	}
}

/*
写消息Goroutine， 用户将数据发送给客户端
*/
func (c *TCPConn) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")
	defer c.Stop()
	for {
		select {
		case data, ok := <-c.writeChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				// break
			}
		case <-c.ctx.Done():
			return
		}
	}
	// for b := range c.writeChan {
	// 	time.Sleep(1 * time.Second)
	// 	fmt.Println("写数据", b)
	// 	if b == nil {
	// 		continue
	// 	}

	// 	_, err := c.conn.Write(b)
	// 	if err != nil {
	// 		break
	// 	}
	// }

}

// (启动连接，让当前连接开始工作)
func (tcpConn *TCPConn) Start() {
	tcpConn.ctx, tcpConn.cancel = context.WithCancel(context.Background())
	go tcpConn.StartReader()
	go tcpConn.StartWriter()
	select {
	case <-tcpConn.ctx.Done():
		tcpConn.finalizer()
		return
	}
}

func (tcpConn *TCPConn) Stop() {
	tcpConn.cancel()
}

func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

// 路由和写数据绑定
func (tcpConn *TCPConn) WriteMsg(msgId uint32, data []byte) error {
	fmt.Println("111")
	tcpConn.RLock()
	defer tcpConn.RUnlock()
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if tcpConn.closeFlag {
		return errors.New("Connection closed when send buff msg")
	}
	pack := msgparser.NewMsgPackage(msgId, data)
	pack.SetMsgId(msgId)
	fmt.Println("222")
	//将data封包，并且发送
	msg, err := tcpConn.msgParser.Encode(pack)
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 发送超时
	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case tcpConn.writeChan <- msg:
		fmt.Println("333")
		return nil
	}
}

// 获取当前连接ID
func (c *TCPConn) GetConnID() uint64 {
	return c.connID
}

// 从当前连接获取原始的socket TCPConn
func (tcpConn *TCPConn) GetTCPConnection() net.Conn {
	return tcpConn.conn
}

func (c *TCPConn) GetWsConn() *websocket.Conn {
	return nil
}

func (tcpConn *TCPConn) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	// c.TCPServer.CallOnConnStop(c)
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}
	tcpConn.closeFlag = true
	tcpConn.conn.Close()
	tcpConn.cancel() //关闭Writer
	//将链接从连接管理器中删除
	tcpConn.TCPServer.GetConnMgr().Remove(tcpConn)
	//关闭该链接全部管道
	close(tcpConn.writeChan)
}
