package acl

import (
	"fmt"
	"sync"
)

type topicSetAuth struct {
	topicM sync.Map
	f      GetAuthFunc
}

var _ Authenticator = (*topicSetAuth)(nil)

func (this *topicSetAuth) CheckPub(userName, topic string) bool {
	return this.CheckSub(userName, topic)
}

func (this *topicSetAuth) CheckSub(userName, topic string) bool {
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	if _, ok := this.topicM.Load(key); ok {
		return true
	}

	exists, ok := this.f(userName, topic).(bool)
	if !ok {
		return false
	}

	if exists {
		this.topicM.Store(key, true)
	}

	return exists
}

func (this *topicSetAuth) ProcessUnSub(userName, topic string) {
	return
}

func (this *topicSetAuth) SetAuthFunc(f GetAuthFunc) {
	this.f = f
}
