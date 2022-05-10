package znet

import "zinx/ziface"

type Requset struct {
	conn ziface.IConnection
	msg  ziface.IMessage
}

// 获取请求连接
func (r *Requset) GetConnection() ziface.IConnection {
	return r.conn
}

// 获取请求数据
func (r *Requset) GetData() []byte {
	return r.msg.GetData()
}

// 获取请求的消息 ID
func (r *Requset) GetMsgID() uint32 {
	return r.msg.GetMsgID()
}
