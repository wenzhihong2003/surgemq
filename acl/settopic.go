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

func (this *topicSetAuth) CheckPub(clientInfo *ClientInfo, topic string) bool {
	return true
}

func (this *topicSetAuth) CheckSub(clientInfo *ClientInfo, topic string) (success bool) {
	token := clientInfo.Token
	key := fmt.Sprintf(userTopicKeyFmt, token, topic)
	if _, ok := this.topicM.Load(key); ok {
		success = true
		return
	}

	var ok bool
	success, ok = this.f(token, topic).(bool)
	if !ok {
		success = false
		return
	}

	if success {
		this.topicM.Store(key, true)
	}

	return
}

func (this *topicSetAuth) ProcessUnSub(clientInfo *ClientInfo, topic string) {
	return
}

func (this *topicSetAuth) SetAuthFunc(f GetAuthFunc) {
	this.f = f
}
