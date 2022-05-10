package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/ziface"
)

// 连接管理模块
type ConnManager struct {
	connections map[uint32]ziface.IConnection // 存储连接信息
	connLock    sync.RWMutex                  // 读写连接锁
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// 添加连接
func (connMgr *ConnManager) Add(conn ziface.IConnection) {
	connMgr.connLock.Lock()
	connMgr.connections[conn.GetConnID()] = conn
	connMgr.connLock.Unlock()
	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}

// 删除连接
func (connMgr *ConnManager) Remove(conn ziface.IConnection) {
	connMgr.connLock.Lock()
	delete(connMgr.connections, conn.GetConnID())
	connMgr.connLock.Unlock()
	fmt.Println("connection Remove ConnID =", conn.GetConnID(), "successfully: conn num =", connMgr.Len())
}

// 获取连接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()
	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

// 获取当前连接数
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

// 清除所有连接
func (connMgr *ConnManager) ClearConn() {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()
	for connID, conn := range connMgr.connections {
		conn.Stop()                         // 停止
		delete(connMgr.connections, connID) // 删除
	}
	fmt.Println("Clear All Connections successfully: conn num =", connMgr.Len())
}
