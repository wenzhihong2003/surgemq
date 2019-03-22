package acl

type alwaysVerify bool

var topicAlwaysVerify alwaysVerify = true

var _ Authenticator = (*alwaysVerify)(nil)

func (this alwaysVerify) CheckPub(clientInfo *ClientInfo, topic string) bool {
	return true

}

func (this alwaysVerify) CheckSub(clientInfo *ClientInfo, topic string) bool {
	operLog("SUB", topic, clientInfo)
	return true

}

func (this alwaysVerify) ProcessUnSub(clientInfo *ClientInfo, topic string) {
	operLog("UNSUB", topic, clientInfo)
}

func (this alwaysVerify) SetAuthFunc(f GetAuthFunc) {

}
