package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"
	"zinx/ziface"
)

// 模拟客户端
func ClientTest() {
	fmt.Println("Client Test Start...")
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("net.Dial err: ", err)
		return
	}
	for {
		// request
		req := NewMsgPackage(0, []byte("ping zinx"))
		pac := NewDataPack()
		data, _ := pac.Pack(req)
		_, err := conn.Write(data)
		if err != nil {
			fmt.Println("conn.Write err: ", err)
			return
		}

		// 服务器回应
		//先读出流中的head部分
		headData := make([]byte, pac.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := pac.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			fmt.Println("==> Recv Msg: ID=", msg.ID, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}
		time.Sleep(time.Second)
	}
}

// 自定义路由 Ping
type PingRouter struct {
	BaseRouter // 继承
}

func (pr *PingRouter) Handle(req ziface.IRequset) {
	fmt.Println("Call PingRouter Handle")
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), req.GetData())

	// 回复
	err := req.GetConnection().SendMsg(1, []byte("Ping... | "))
	if err != nil {
		fmt.Println("PingRouter err", err)
	}
}

// Hello
type HelloRouter struct {
	BaseRouter
}

func (hr *HelloRouter) Handle(req ziface.IRequset) {
	fmt.Println("Call HelloRouter Handle")
	// 读取客户端数据
	fmt.Printf("recv from client: [msgID=%d] [data=%s]\n", req.GetMsgID(), req.GetData())

	// 回复
	err := req.GetConnection().SendMsg(1, []byte("Hello... | "))
	if err != nil {
		fmt.Println("HelloRouter err", err)
	}
}

// Server 模块测试函数
func TestServe(t *testing.T) {
	// 服务端测试
	s := NewServer()
	// 注册路由
	s.AddRouter(0, new(PingRouter))
	s.AddRouter(1, new(HelloRouter))
	// 客户端测试
	go ClientTest()
	// 开启服务
	s.Serve()
}
