package network

import (
	"errors"
	"fmt"
	"gobonbon/iface"
	"sync"
)

/*
连接管理模块
*/
type ConnManager struct {
	connSet  map[uint64]iface.IConn //管理的连接信息
	connLock sync.RWMutex           //读写连接的读写锁
}

/*
创建一个链接管理
*/
func NewConnManager() *ConnManager {
	return &ConnManager{
		connSet: make(map[uint64]iface.IConn),
	}
}

// 添加链接
func (connMgr *ConnManager) Add(conn iface.IConn) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	//将conn连接添加到ConnMananger中
	connMgr.connSet[conn.GetConnID()] = conn
	connMgr.connLock.Unlock()
	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}

// 删除连接
func (connMgr *ConnManager) Remove(conn iface.IConn) {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	//删除连接信息
	delete(connMgr.connSet, conn.GetConnID())
	connMgr.connLock.Unlock()
	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
}

// 利用ConnID获取链接
func (connMgr *ConnManager) Get(connID uint64) (iface.IConn, error) {
	//保护共享资源Map 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connSet[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

// 获取当前连接
func (connMgr *ConnManager) Len() int {
	connMgr.connLock.RLock()
	length := len(connMgr.connSet)
	connMgr.connLock.RUnlock()
	return length
}

// 清除并停止所有连接
func (connMgr *ConnManager) ClearConn() {
	//保护共享资源Map 加写锁
	connMgr.connLock.Lock()
	//停止并删除全部的连接信息
	for _, conn := range connMgr.connSet {
		//停止
		conn.Stop()
	}
	connMgr.connLock.Unlock()
	fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
}

// ClearOneConn  利用ConnID获取一个链接 并且删除
func (connMgr *ConnManager) ClearOneConn(connID uint64) {
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	connections := connMgr.connSet
	if conn, ok := connections[connID]; ok {
		//停止
		conn.Stop()

		fmt.Println("Clear Connections ID:  ", connID, "succeed")
		return
	}

	fmt.Println("Clear Connections ID:  ", connID, "err")
}
