package socket

import (
	"sync"
	"sync/atomic"

	"github.com/davyxu/cellnet"
)

type SessionManager struct {
	sesMap map[int64]cellnet.Session

	sesIDAcc    int64
	sesMapGuard sync.RWMutex
}

const totalTryCount = 100

func (self *SessionManager) Add(ses cellnet.Session) {

	self.sesMapGuard.Lock()
	defer self.sesMapGuard.Unlock()

	var tryCount int = totalTryCount

	var id int64

	// id翻越处理
	for tryCount > 0 {

		id = atomic.AddInt64(&self.sesIDAcc, 1)

		if _, ok := self.sesMap[id]; !ok {
			break
		}

		tryCount--
	}

	if tryCount == 0 {
		log.Warnln("sessionID override!", id)
	}

	socketSes := ses.(*SocketSession)

	socketSes.id = id

	self.sesMap[id] = ses

}

func (self *SessionManager) Remove(ses cellnet.Session) {
	self.sesMapGuard.Lock()
	delete(self.sesMap, ses.ID())
	self.sesMapGuard.Unlock()
}

// 获得一个连接
func (self *SessionManager) GetSession(id int64) cellnet.Session {
	self.sesMapGuard.RLock()
	defer self.sesMapGuard.RUnlock()

	v, ok := self.sesMap[id]
	if ok {
		return v
	}

	return nil
}

func (self *SessionManager) VisitSession(callback func(cellnet.Session) bool) {
	self.sesMapGuard.RLock()
	defer self.sesMapGuard.RUnlock()

	for _, ses := range self.sesMap {
		if !callback(ses) {
			break
		}
	}

}

func (self *SessionManager) CloseAllSession() {

	self.VisitSession(func(ses cellnet.Session) bool {

		ses.Close()

		return true
	})
}

func (self *SessionManager) SessionCount() int {
	self.sesMapGuard.Lock()
	defer self.sesMapGuard.Unlock()

	return len(self.sesMap)
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sesMap: make(map[int64]cellnet.Session),
	}
}

var ClientSessionManager = NewSessionManager()
