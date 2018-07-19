  1.新增特性:
 * **auth** （client身份信息校验以及获取）
 * **acl**  （topic的订阅时对用户进行限制）

  2.使用方法
  * auth
 
       
            type mockAuthenticator bool

            var _ Authenticator = (*mockAuthenticator)(nil)

            var (
                mockSuccessAuthenticator mockAuthenticator = true
                mockFailureAuthenticator mockAuthenticator = false
            )

            func init() {
                Register("mockSuccess", mockSuccessAuthenticator)
                Register("mockFailure", mockFailureAuthenticator)
            }

            func (this mockAuthenticator) Authenticate(token, userName string) (bool, *ClientInfo) {

                return this == mockSuccessAuthenticator, &ClientInfo{Token: token}
            }

            func (this mockAuthenticator) SetAuthFunc(f AuthFunc) {

            }
  * acl 


            type alwaysVerify bool

            var topicAlwaysVerify alwaysVerify = true

            var _ Authenticator = (*alwaysVerify)(nil)

            func (this alwaysVerify) CheckPub(clientInfo *ClientInfo, topic string) bool {
                return true

            }

            func (this alwaysVerify) CheckSub(clientInfo *ClientInfo, topic string) bool {
                return true

            }

            func (this alwaysVerify) ProcessUnSub(clientInfo *ClientInfo, topic string) {
            }

            func (this alwaysVerify) SetAuthFunc(f GetAuthFunc) {

            }


