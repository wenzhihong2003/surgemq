package acl

import (
	"errors"
)

type Authenticator interface {
	CheckPub(topic []byte) bool
	CheckSub(topic []byte) bool
	ProcessUnSub(topic []byte)
}

type TopicAclManger struct {
	p Authenticator
}

func (this *TopicAclManger) CheckPub(topic []byte) bool {
	return this.p.CheckPub(topic)
}

func (this *TopicAclManger) CheckSub(topic []byte) bool {
	return this.p.CheckSub(topic)
}

func (this *TopicAclManger) ProcessUnSub(topic []byte) {
	this.p.ProcessUnSub(topic)
	return
}

func NewTopicAclManger(providerName string, authInfo interface{}) (*TopicAclManger, error) {

	switch providerName {
	case TopicNumAuthType:
		return &TopicAclManger{NewTopicNumAuthProvider(authInfo)}, nil

	case TopicSetAuthType:
		return &TopicAclManger{NewTopicSetAuthProvider(authInfo)}, nil

	case TopicAlwaysVerifyType:
		var yes TopicAlwaysVerify
		return &TopicAclManger{yes}, nil

	default:
		return nil, errors.New("not exists provider:" + providerName)

	}

}
