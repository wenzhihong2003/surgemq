package acl

import (
	"fmt"
	"sync"
)

type topicSetAuth struct {
	topicM sync.Map
	f      GetAuthFunc // must return bool
}

var _ Authenticator = (*topicSetAuth)(nil)

func (this *topicSetAuth) CheckPub(clientInfo *ClientInfo, topic string) bool {
	return true
}

func (this *topicSetAuth) CheckSub(clientInfo *ClientInfo, topic string) (success bool) {
	if clientInfo == nil {
		return
	}
	token := clientInfo.GmToken
	key := fmt.Sprintf(userTopicKeyFmt, token, topic)
	if v, ok := this.topicM.Load(key); ok {
		success = v.(bool)
		return
	}

	success = this.f(clientInfo, topic).(bool)
	this.topicM.Store(key, success)

	return
}

func (this *topicSetAuth) ProcessUnSub(clientInfo *ClientInfo, topic string) {
	return
}

func (this *topicSetAuth) SetAuthFunc(f GetAuthFunc) {
	this.f = f
}
