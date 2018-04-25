package acl

import (
	"errors"
)

const (
	TopicAlwaysVerifyType = "topicAlwaysVerify"
	TopicNumAuthType      = "topicNumAuth"
	TopicSetAuthType      = "topicSetAuth"
	userTopicKeyFmt       = "%s:%s"
)

type GetAuthFunc func(userName, topic string) interface{}

type Authenticator interface {
	CheckPub(userName, topic string) bool
	CheckSub(userName, topic string) bool
	ProcessUnSub(userName, topic string)
	SetAuthFunc(f GetAuthFunc)
}

var providers = make(map[string]Authenticator)

type TopicAclManger struct {
	p Authenticator
}

func (this *TopicAclManger) CheckPub(userName, topic string) bool {
	return this.p.CheckPub(userName, topic)
}

func (this *TopicAclManger) CheckSub(userName, topic string) bool {
	return this.p.CheckSub(userName, topic)
}

func (this *TopicAclManger) ProcessUnSub(userName, topic string) {
	this.p.ProcessUnSub(userName, topic)
	return
}

func (this *TopicAclManger) SetAuthFunc(f GetAuthFunc) {
	this.p.SetAuthFunc(f)
}

func NewTopicAclManger(providerName string, f GetAuthFunc) (*TopicAclManger, error) {
	v, ok := providers[providerName]
	if !ok {
		return nil, errors.New("providers not exist this name:" + providerName)
	}
	topicManger := &TopicAclManger{v}
	topicManger.SetAuthFunc(f)
	return topicManger, nil
}

func Register(name string, provider Authenticator) {
	if provider == nil || len(name) == 0 {
		panic("传入参数name和provider有误!")
	}

	if _, dup := providers[name]; dup {
		panic("auth: Register called twice for provider " + name)
	}

	providers[name] = provider
}

func UnRegister(name string) {
	delete(providers, name)
}

func init() {
	Register(TopicAlwaysVerifyType, topicAlwaysVerify)
	Register(TopicNumAuthType, new(topicNumAuth))
	Register(TopicSetAuthType, new(topicSetAuth))
}
