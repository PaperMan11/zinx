package znet

import "zinx/ziface"

type Message struct {
	ID      uint32 // 消息ID
	DataLen uint32 // 消息长度
	Data    []byte // 消息内容
}

// 创建一个消息对象
func NewMsgPackage(id uint32, data []byte) ziface.IMessage {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// 获取数据信息
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

func (msg *Message) GetMsgID() uint32 {
	return msg.ID
}

func (msg *Message) GetData() []byte {
	return msg.Data
}

// 设置数据信息
func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

func (msg *Message) SetMsgID(id uint32) {
	msg.ID = id
}

func (msg *Message) SetData(data []byte) {
	msg.Data = data
}
