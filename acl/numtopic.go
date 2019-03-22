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

func (this *topicNumAuth) CheckPub(clientInfo *ClientInfo, topic string) bool {
	return true
}

// 需要区分tick和bar
func (this *topicNumAuth) CheckSub(clientInfo *ClientInfo, topic string) (success bool) {

	if clientInfo == nil {
		return false
	}

	defer func() {
		operLog("SUB", topic, clientInfo)
	}()

	// fmt.Println("clientInfo",*clientInfo)
	userName := clientInfo.GmToken
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	if _, ok := this.topicUserM.Load(key); ok {
		success = true
		return true
	}

	totalLimit := clientInfo.SubTopicLimit

	totalNow, ok := this.topicTotalNowM.Load(userName)
	if !ok {
		this.topicTotalNowM.Store(userName, 1)
		this.topicUserM.Store(key, true)
		success = true
		return
	}

	if totalNow.(int) >= totalLimit {
		return
	}

	this.topicTotalNowM.Store(userName, totalNow.(int)+1)
	this.topicUserM.Store(key, true)
	success = true
	return
}

func (this *topicNumAuth) ProcessUnSub(clientInfo *ClientInfo, topic string) {
	if clientInfo == nil {
		return
	}

	defer func() {
		operLog("UNSUB", topic, clientInfo)
	}()

	userName := clientInfo.GmToken
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
