package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

// 连接对象
type Connection struct {
	TcpServer    ziface.IServer         // 当前连接属于哪个 Server
	Conn         *net.TCPConn           // 连接套接字
	ConnID       uint32                 // 连接ID
	isClosed     bool                   // 是否关闭连接
	MsgHandler   ziface.IMsgHandle      // 消息管理MsgId和对应处理方法的消息管理模块
	ExitBuffChan chan struct{}          // 告知连接关闭的channel
	msgChan      chan []byte            // 用于读、写两个 groutine 之间的消息通信
	msgBuffChan  chan []byte            // 有缓冲
	property     map[string]interface{} // 连接属性
	propertyLock sync.RWMutex           // 锁
}

// 创建连接方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}
	// 将新创建的 conn 加入 connMgr 中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// 处理 conn 读取数据的 Groutine
func (c *Connection) StartReader() {
	fmt.Println("Reader Groutine runnig...")
	defer fmt.Println(c.Conn.RemoteAddr().String(), "conn reader exit...")
	defer c.Stop()

	for {
		// 1、读取客户端请求包
		// 创建包对象
		dp := NewDataPack()

		// 1.1、读取包的 Head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("Read msg head err: ", err)
			c.ExitBuffChan <- struct{}{}
			break
		}

		// 1.2、拆包得到 msgID msgLen
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("Unpack err: ", err)
			c.ExitBuffChan <- struct{}{}
			break
		}

		// 1.3、继续读取包的 data
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("Read msg data err: ", err)
				c.ExitBuffChan <- struct{}{}
				continue
			}
		}
		msg.SetData(data) // 具体数据内容放入msg中

		// 1.4、设置 req 数据
		req := Requset{
			conn: c,
			msg:  msg,
		}

		// 2、从绑定好的消息和对应的处理方法中执行对应的 Handle 方法
		if utils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req) // 将请求加入任务队列，对应的协程进行处理
		} else {
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

// 处理 conn 写数据的 Groutine
func (c *Connection) StartWriter() {
	fmt.Println("Writer Groutine is running...")
	defer fmt.Println(c.RemoteAddr().String(), "conn Writer exit...")

	for {
		select {
		case data := <-c.msgChan: // 无缓冲
			// 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data err: ", err)
				return
			}
		case data, ok := <-c.msgBuffChan: // 有缓冲
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data err: ", err)
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				break
			}
		case <-c.ExitBuffChan:
			// conn 关闭
			return
		}
	}
}

// 启动连接
func (c *Connection) Start() {
	go c.StartReader()             // 读客户端数据
	go c.StartWriter()             // 写客户端数据
	c.TcpServer.CallOnConnStart(c) // 运行 hook 函数 (!nil)
	for {
		select {
		case <-c.ExitBuffChan:
			return
		}
	}
}

// 关闭连接
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()...ConnID = ", c.ConnID)
	// 当前连接已关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	// 运行 hook 函数 (!nil)
	c.TcpServer.CallOnConnStop(c)

	// 关闭当前连接
	c.Conn.Close()
	c.ExitBuffChan <- struct{}{}

	// 将连接从连接管理对象中删除
	c.TcpServer.GetConnMgr().Remove(c)

	// 关闭所有管道
	close(c.ExitBuffChan)
	close(c.msgChan)
	close(c.msgBuffChan)
}

// 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 回复客户端 (无缓冲)
func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}
	// 包对象
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgID, data))
	if err != nil {
		fmt.Println("Pack error msg id: ", msgID)
		return errors.New("Pack error msg")
	}

	// 回复客户端 (读写分离)
	// Write
	c.msgChan <- msg
	return nil
}

// 回复客户端 (有缓冲)
func (c *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	// 包对象
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgID, data))
	if err != nil {
		fmt.Println("Pack error msg id: ", msgID)
		return errors.New("Pack error msg")
	}
	// 回复客户端 (读写分离)
	// Write
	c.msgBuffChan <- msg
	return nil
}

// 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

// 获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// 移除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}
