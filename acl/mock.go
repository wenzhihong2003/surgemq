package acl
type TopicAlwaysVerify bool

var _ Authenticator = (*TopicAlwaysVerify)(nil)

func (this TopicAlwaysVerify) CheckPub(topic []byte) bool {
	return true

}

func (this TopicAlwaysVerify) CheckSub(topic []byte) bool {
	return true

}

func (this TopicAlwaysVerify) ProcessUnSub(topic []byte) {
}
