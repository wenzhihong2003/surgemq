package acl

import (
	"fmt"
	"sync"
)

const (
	TopicAlwaysVerifyType = "topicAlwaysVerify"
	TopicNumAuthType      = "topicNumAuth"
	TopicSetAuthType      = "topicSetAuth"
	userTopicKeyFmt       = "%s:%s"
)

type TopicAlwaysVerify bool

var topicAlwaysVerify TopicAlwaysVerify = true

var _ Authenticator = (*TopicAlwaysVerify)(nil)

func init() {
	Register(TopicAlwaysVerifyType, topicAlwaysVerify)
	Register(TopicNumAuthType, new(topicNumAuth))
	Register(TopicSetAuthType, new(topicSetAuth))
}

func (this TopicAlwaysVerify) CheckPub(userName, topic string) bool {
	return true

}

func (this TopicAlwaysVerify) CheckSub(userName, topic string) bool {
	return true

}

func (this TopicAlwaysVerify) ProcessUnSub(userName, topic string) {
}

type topicNumAuth struct {
	topicTotalNowM sync.Map
	topicUserM     sync.Map
}

var _ Authenticator = (*topicNumAuth)(nil)

func (this *topicNumAuth) CheckPub(userName, topic string) bool {
	return true
}

func (this *topicNumAuth) CheckSub(userName, topic string) bool {
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	totalLimit, ok := getAuth(userName, topic).(int)
	if !ok {
		return false
	}

	if totalLimit == 0 {
		return false
	}

	totalNow, ok := this.topicTotalNowM.Load(userName)
	if !ok {
		this.topicTotalNowM.Store(userName, 1)
		this.topicUserM.Store(key, true)
		return true
	}

	if totalNow.(int) > totalLimit {
		return false
	}

	if _, ok := this.topicUserM.Load(key); ok {
		return true
	}

	if totalNow.(int) == totalLimit {
		return false
	}

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

type topicSetAuth struct {
	topicM sync.Map
}

var _ Authenticator = (*topicSetAuth)(nil)

func (this *topicSetAuth) CheckPub(userName, topic string) bool {
	return true
}

func (this *topicSetAuth) CheckSub(userName, topic string) bool {
	key := fmt.Sprintf(userTopicKeyFmt, userName, topic)
	if _, ok := this.topicM.Load(key); ok {
		return true
	}

	exists, ok := getAuth(userName, topic).(bool)
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
