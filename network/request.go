package network

import (
	"fmt"
	"gobonbon/iface"
	"sync"
)

const (
	PRE_HANDLE  iface.HandleStep = iota //PreHandle 预处理  0
	HANDLE                              //Handle 处理
	POST_HANDLE                         //PostHandle 后处理
	HANDLE_OVER
)

type Request struct {
	conn     iface.IConn      //已经和客户端建立好的 链接
	msg      iface.IMessage   //客户端请求的数据
	router   iface.IRouter    //请求处理的函数
	steps    iface.HandleStep //用来控制路由函数执行
	stepLock *sync.RWMutex    //并发互斥
	needNext bool             //是否需要执行下一个路由函数
}

func NewRequest(conn iface.IConn, msg iface.IMessage) *Request {
	req := new(Request)
	req.steps = PRE_HANDLE
	req.conn = conn
	req.msg = msg
	req.stepLock = new(sync.RWMutex)
	req.needNext = true

	return req
}

// 获取请求连接信息
func (r *Request) GetConnection() iface.IConn {
	return r.conn
}

// 获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

// GetMsgID implements IRequest.
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}

func (r *Request) BindRouter(router iface.IRouter) {
	r.router = router
}

func (r *Request) next() {
	if !r.needNext {
		r.needNext = true
		return
	}
	r.stepLock.Lock()
	r.steps++
	r.stepLock.Unlock()
}

func (r *Request) Call() {
	if r.router == nil {
		return
	}
	fmt.Printf(" server , name %d", 111)
	for r.steps < HANDLE_OVER {
		switch r.steps {
		case PRE_HANDLE:
			r.router.PreHandle(r)
		case HANDLE:
			r.router.Handle(r)
		case POST_HANDLE:
			r.router.PostHandle(r)
		}
		r.next()
	}
}

func (r *Request) Abort() {
	r.stepLock.Lock()
	r.steps = HANDLE_OVER
	r.stepLock.Unlock()
}

func (r *Request) Goto(step iface.HandleStep) {
	r.stepLock.Lock()
	r.steps = step
	r.needNext = false
	r.stepLock.Unlock()
}
