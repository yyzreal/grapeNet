// 存放数据结构
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/7/12

package grapeConn

import (
	"sync"

	logger "github.com/koangel/grapeNet/Logger"
)

const (
	ESERVER_TYPE = iota
	ECLIENT_TYPE
)

const (
	defaultChan = 1024
)

type ConnInterface interface {
	GetSessionId() string
	Send(data []byte) int
	SendPak(val interface{}) int
	Close()

	RemoveData()
	InitData()

	startProc()
}

type Conn struct {
	SessionId string
	Type      int
	Done      chan int
}

func (c *Conn) GetSessionId() string {
	return c.SessionId
}

func (c *Conn) Send(data []byte) int {
	return -1
}

func (c *Conn) Close() {

}

func (c *Conn) RemoveData() {

}

func (c *Conn) InitData() {

}

func (c *Conn) startProc() {

}

func (c *Conn) SendPak(val interface{}) int {
	return -1
}

type ConnManager struct {
	continer map[ConnInterface]bool   // 存放主要数据
	sessions map[string]ConnInterface // 查询SESSION

	Register   chan ConnInterface
	Unregister chan ConnInterface

	locker sync.RWMutex // 锁
}

func NewCM() *ConnManager {
	newCm := &ConnManager{
		continer:   make(map[ConnInterface]bool),
		sessions:   make(map[string]ConnInterface),
		Register:   make(chan ConnInterface, defaultChan),
		Unregister: make(chan ConnInterface, defaultChan),
	}

	go newCm.process()

	return newCm
}

func (c *ConnManager) process() {
	defer func() {
		logger.TRACE("Conn Manager Closed...")
	}()

	for {
		select {
		case conn, rok := <-c.Register:
			if !rok {
				return
			}

			logger.TRACE("Register In Conn -> %v...", conn.GetSessionId())
			// 加入map
			c.locker.Lock()
			c.continer[conn] = true
			c.sessions[conn.GetSessionId()] = conn
			c.locker.Unlock()

			conn.InitData() // 初始化数据
			break
		case conn, rok := <-c.Unregister:
			if !rok {
				return
			}

			logger.TRACE("Unregister In Conn -> %v", conn.GetSessionId())
			conn.Close()

			// 加入map
			c.locker.Lock()
			delete(c.continer, conn)
			delete(c.sessions, conn.GetSessionId())
			c.locker.Unlock()

			conn.RemoveData() // 移除数据

			break
		}
	}
}

func (c *ConnManager) Get(sessionId string) ConnInterface {
	c.locker.RLock()
	defer c.locker.RUnlock()

	val, ok := c.sessions[sessionId]
	if !ok {
		return nil
	}

	return val
}

func (c *ConnManager) Broadcast(data []byte) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	for _, v := range c.sessions {
		v.Send(data)
	}
}

func (c *ConnManager) BroadcastExcep(sessionId string, data []byte) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	for k, v := range c.sessions {
		if k == sessionId {
			continue
		}

		v.Send(data)
	}
}
