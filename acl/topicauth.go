package acl

import (
	"github.com/pkg/errors"
)

type GetAuthFunc func(userName, topic string) interface{}

var getAuth GetAuthFunc //callback get authInfo of user

type Authenticator interface {
	CheckPub(userName, topic string) bool
	CheckSub(userName, topic string) bool
	ProcessUnSub(userName, topic string)
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

func NewTopicAclManger(providerName string, f GetAuthFunc) (*TopicAclManger, error) {
	getAuth = f
	v, ok := providers[providerName]
	if !ok {
		return nil, errors.New("providers not exist this name:" + providerName)
	}

	return &TopicAclManger{v}, nil
}

var DefalutNewTopicAclMangerFunc = func(userName string) (*TopicAclManger, error) {
	var yes TopicAlwaysVerify
	return &TopicAclManger{yes}, nil
}

func Register(name string, provider Authenticator) {
	if provider == nil {
		panic("auth: Register provide is nil")
	}

	if _, dup := providers[name]; dup {
		panic("auth: Register called twice for provider " + name)
	}

	providers[name] = provider
}

func UnRegister(name string) {
	delete(providers, name)
}
