package znet

import (
	"fmt"
	"net"
	"time"
	"zinx/utils"
	"zinx/ziface"
)

// iServer 接口实现，定义一个 Server 服务类
type Server struct {
	Name      string              // 服务器名称
	IPVersion string              // tcp4 or other
	IP        string              // 服务绑定的IP
	Port      int                 // 服务绑定的端口
	MsgHandle ziface.IMsgHandle   // 当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	ConnMgr   ziface.IConnManager // 连接管理

	// 两个 Hook 函数原型
	OnConnStart func(conn ziface.IConnection) // 连接启动时运行的函数
	OnConnStop  func(conn ziface.IConnection) // 连接结束时运行的函数
}

// 实现接口的方法
func (s *Server) Start() {
	fmt.Printf("[START] Server <Name: %s>, Listenner <IP: %s, Port: %d>\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)
	// Listen AND Accept
	go func() {
		// 0、启动工作池
		s.MsgHandle.StartWorkerPool()

		// 1、获取一个 TCP 的 Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("net.ResolveTCPAddr err: ", err)
			return
		}

		// 2、监听服务器地址
		lis, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("net.ListenTCP err: ", err)
			return
		}
		fmt.Println("start Zinx server", s.Name, "succ, now listenning...")

		//TODO server.go 应该有一个自动生成ID的方法
		var cid uint32 = 0

		// 3、启动网络连接业务
		for {
			conn, err := lis.AcceptTCP()
			if err != nil {
				fmt.Println("lis.AcceptTCP err: ", err)
				continue
			}
			// 设置最大连接数控制
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}
			// 连接处理
			dealConn := NewConnection(s, conn, cid, s.MsgHandle)
			cid++
			go dealConn.Start() // 处理连接业务
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)
	//TODO  Server.Stop() 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()
	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listenner的go将会退出
	for {
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// 设置 hook 函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用 hook 函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("----> CallOnConnStart <----")
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("----> CallOnConnStop <----")
		s.OnConnStop(conn)
	}
}

// 创建一个服务器句柄
func NewServer() ziface.IServer {
	// 初始化全局配置文件
	utils.GlobalObject.Reload()
	return &Server{
		Name:      utils.GlobalObject.Name, // 全局参数获取
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}
}
