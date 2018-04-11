package acl

type Authenticator interface {
	CheckPub(topic []byte) bool
	CheckSub(topic []byte) bool
	ProcessUnSub(topic []byte)
}

type TopicAclManger struct {
	P Authenticator
}

func (this *TopicAclManger) CheckPub(topic []byte) bool {
	return this.P.CheckPub(topic)
}

func (this *TopicAclManger) CheckSub(topic []byte) bool {
	return this.P.CheckSub(topic)
}

func (this *TopicAclManger) ProcessUnSub(topic []byte) {
	this.P.ProcessUnSub(topic)
	return
}

type NewTopicAclMangerFunc func(userName string) (*TopicAclManger, error)

var DefalutNewTopicAclMangerFunc = func(userName string) (*TopicAclManger, error) {
	var yes TopicAlwaysVerify
	return &TopicAclManger{yes}, nil
}
