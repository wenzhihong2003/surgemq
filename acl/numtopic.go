package acl

import (
	"fmt"
	"sync"
)

type topicNumAuth struct {
	topicTotalNowM sync.Map
	topicUserM     sync.Map
	f              GetAuthFunc
}

var _ Authenticator = (*topicNumAuth)(nil)

func (this *topicNumAuth) CheckPub(userName, topic string) bool {
	return true
}

func (this *topicNumAuth) CheckSub(userName, topic string) bool {
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	if _, ok := this.topicUserM.Load(key); ok {
		return true
	}

	totalLimit, ok := this.f(userName, topic).(int)
	if !ok || totalLimit == 0 {
		return false
	}

	totalNow, ok := this.topicTotalNowM.Load(userName)
	if !ok {
		this.topicTotalNowM.Store(userName, 1)
		this.topicUserM.Store(key, true)
		return true
	}

	if totalNow.(int) >= totalLimit {
		return false
	}

	this.topicTotalNowM.Store(userName, totalNow.(int)+1)
	this.topicUserM.Store(key, true)
	return true
}

func (this *topicNumAuth) ProcessUnSub(userName, topic string) {
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	if _, ok := this.topicUserM.Load(key); !ok {
		return
	}
	this.topicUserM.Delete(key)
	totalNow, ok := this.topicTotalNowM.Load(userName)
	if ok {
		this.topicTotalNowM.Store(userName, totalNow.(int)-1)
	}
}

func (this *topicNumAuth) SetAuthFunc(f GetAuthFunc) {
	this.f = f
}
