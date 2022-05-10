package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/znet"
)

func main() {
	fmt.Println("Client_11 Test Start...")
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("net.Dial err: ", err)
		return
	}
	for {
		// request
		req := znet.NewMsgPackage(0, []byte("ping zinx"))
		pac := znet.NewDataPack()
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
			msg := msgHead.(*znet.Message)
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
