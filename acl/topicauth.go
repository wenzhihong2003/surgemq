package acl

import (
	"errors"
)

type Authenticator interface {
	CheckPub(topic string) bool
	CheckSub(topic string) bool
	ProcessUnSub(topic string)
}

type AuthInfo struct {
	Total    int
	TopicMap map[string]bool
}

type TopicAclManger struct {
	p Authenticator
}

func (this *TopicAclManger) CheckPub(topic string) bool {
	return this.p.CheckPub(topic)
}

func (this *TopicAclManger) CheckSub(topic string) bool {
	return this.p.CheckSub(topic)
}

func (this *TopicAclManger) ProcessUnSub(topic string) {
	this.p.ProcessUnSub(topic)
	return
}

func NewTopicAclManger(providerName string, authInfo *AuthInfo) (*TopicAclManger, error) {

	switch providerName {
	case TopicNumAuthType:
		return &TopicAclManger{NewTopicNumAuthProvider(authInfo.Total)}, nil

	case TopicSetAuthType:
		return &TopicAclManger{NewTopicSetAuthProvider(authInfo.TopicMap)}, nil

	case TopicAlwaysVerifyType:
		var yes TopicAlwaysVerify
		return &TopicAclManger{yes}, nil

	default:
		return nil, errors.New("not exists provider:" + providerName)

	}

}
