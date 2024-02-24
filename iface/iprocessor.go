package iface

type HandleStep int

// 定义服务器接口
type IServer interface {
	Start() //启动服务器方法
	Stop()  //停止服务器方法
	// Serve()//开启业务服务方法

	AddRouter(msgId uint32, router IRouter) //路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	GetConnMgr() IConnManager               //得到链接管理

	ServerName() string        // Get the server name (获取服务器名称)
	GetMsgHandler() IMsgHandle // (获取Server绑定的消息处理模块)
}

// 实际上是把客户端请求的链接信息 和 请求的数据 包装到了 Request里，利于之后拓展框架
// 可以理解为每次客户端的全部请求数据，都会把它们一起放到一个Request结构体里
type IRequest interface {
	GetConnection() IConn      //获取请求连接信息
	GetData() []byte           //获取请求消息的数据
	GetMsgID() uint32          // Get the message ID of the request(获取请求的消息ID)
	BindRouter(router IRouter) //绑定这次请求由哪个路由处理
	Call()                     //转进到下一个处理器开始执行 但是调用此方法的函数会根据先后顺序逆序执行
	Abort()                    //终止处理函数的运行 但调用此方法的函数会执行完毕
	//慎用，会导致循环调用
	Goto(HandleStep) //指定接下来的Handle去执行哪个Handler函数
}

// 路由接口， 这里面路由是 使用框架者给该链接自定的 处理业务方法
// 路由里的IRequest 则包含用该链接的链接信息和该链接的请求数据信息
type IRouter interface {
	PreHandle(request IRequest)  //在处理conn业务之前的钩子方法
	Handle(request IRequest)     //处理conn业务的方法
	PostHandle(request IRequest) //处理conn业务之后的钩子方法
}

// 消息管理抽象层
type IMsgHandle interface {
	DoMsgHandler(request IRequest)          //马上以非阻塞方式处理消息
	AddRouter(msgId uint32, router IRouter) //为消息添加具体的处理逻辑
	StartWorkerPool()                       //启动worker工作池
	SendMsgToTaskQueue(request IRequest)    //将消息交给TaskQueue,由worker进行处理
}

// 将TCP请求的一个消息封装到message中，定义抽象层接口
type IMsgParser interface {
	GetHeadLen() uint32                  //获取包头长度方法
	Encode(msg IMessage) ([]byte, error) //封包方法
	Decode([]byte) (IMessage, error)     //拆包方法
}

type IMessage interface {
	GetDataLen() uint32 //获取消息数据段长度
	GetMsgId() uint32   //获取消息ID
	GetData() []byte    //获取消息内容

	SetMsgId(uint32)   //设计消息ID
	SetData([]byte)    //设计消息内容
	SetDataLen(uint32) //设置消息数据段长度
}

// 连接管理抽象层
type IConnManager interface {
	Add(conn IConn)                   //添加链接
	Remove(conn IConn)                //删除连接
	Get(connID uint64) (IConn, error) //利用ConnID获取链接
	Len() int                         //获取当前连接
	ClearConn()                       //删除并停止所有链接
}
