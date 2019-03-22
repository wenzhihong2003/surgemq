调用 `service.InitParamFromEnv()` 可以从环境变量读取值初始化 inBuffer, outBuffer 及 pubMsgChanSize 值, 主要用于默认的参数不符合我们应用场景的情况


#### gInBufferSize
用于mqtt server接收客户端发来的消息,
默认为: 256 * 1024, 也就是256k, 可通过设置环境变量: `GM_IN_BUFFER_SIZE` 来设置. 设置为 128, 256, 512, 1024 这样类似的值

#### gOutBufferSize
用于mqtt server往客户端发送消息
默认为: 256 * 1024, 也就是256k, 可通过设置环境变量: `GM_OUT_BUFFER_SIZE` 来设置. 设置为 128, 256, 512, 1024 这样类似的值

#### gPubMsgChanSize
用于mqtt server给客户端的每个连接发送消息的队列大小
默认为: 300, 可通过环境变量 `GM_PUB_MSG_CHAN_SIZE` 来设置. 最大值为: 1024 * 50 = 51200


#### 新增特性:
 * **auth** （client身份信息校验以及获取）
 * **acl**  （topic的订阅时对用户进行限制）

#### 使用方法

* auth

 ```
       
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
 ```


* acl
```
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
```

