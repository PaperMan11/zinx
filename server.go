package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

// 自定义路由 Ping
type PingRouter struct {
	znet.BaseRouter // 继承
}

func (pr *PingRouter) Handle(req ziface.IRequset) {
	fmt.Println("Call PingRouter Handle")
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), req.GetData())

	// 回复
	err := req.GetConnection().SendMsg(0, []byte("Ping... | "))
	if err != nil {
		fmt.Println("PingRouter err", err)
	}
}

// Hello
type HelloRouter struct {
	znet.BaseRouter
}

func (hr *HelloRouter) Handle(req ziface.IRequset) {
	fmt.Println("Call HelloRouter Handle")
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), req.GetData())

	// 回复
	err := req.GetConnection().SendBuffMsg(1, []byte("Hello... | "))
	if err != nil {
		fmt.Println("HelloRouter err", err)
	}
}

// 设置 hook 函数
// 创建连接时执行
func DoConnectionStart(conn ziface.IConnection) {
	//=============设置两个链接属性，在连接创建之后===========
	fmt.Println("Set conn Name, Home done!")
	conn.SetProperty("Name", "tzq")
	conn.SetProperty("Home", "github.com")
	//===================================================
	fmt.Println("<<< conn start hook func >>>")
}

// 断开连接时执行
func DoConnectionStop(conn ziface.IConnection) {
	//============在连接销毁之前，查询conn的Name，Home属性=====
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		fmt.Println("Conn Property Home = ", home)
	}
	//===================================================
	fmt.Println("<<< conn stop hook func >>>")
}

func main() {
	// 服务端测试
	s := znet.NewServer()
	// 注册路由
	s.AddRouter(0, new(PingRouter))
	s.AddRouter(1, new(HelloRouter))
	// 设置 hook 函数
	s.SetOnConnStart(DoConnectionStart)
	s.SetOnConnStop(DoConnectionStop)

	// 开启服务
	s.Serve()
}
