package acl

type alwaysVerify bool

var topicAlwaysVerify alwaysVerify = true

var _ Authenticator = (*alwaysVerify)(nil)

func (this alwaysVerify) CheckPub(userName, topic string) bool {
	return true

}

func (this alwaysVerify) CheckSub(userName, topic string) bool {
	return true

}

func (this alwaysVerify) ProcessUnSub(userName, topic string) {
}

func (this alwaysVerify) SetAuthFunc(f GetAuthFunc) {

}
