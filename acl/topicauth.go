package acl

import (
	"errors"

	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TopicAlwaysVerifyType = "topicAlwaysVerify"
	TopicNumAuthType      = "topicNumAuth"
	TopicSetAuthType      = "topicSetAuth"
	userTopicKeyFmt       = "%s:%s"
)

type GetAuthFunc func(userName, topic string) interface{}

type ClientInfo struct {
	Token    string
	UserName string
	UserId   string
	SdkInfo  *SdkInfo
}

//"sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64",
type SdkInfo struct {
	SdkLang    string
	SdkVersion string
	SdkArch    string
	SdkOs      string
}

//sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64
type Authenticator interface {
	CheckPub(clientInfo *ClientInfo, topic string) bool
	CheckSub(clientInfo *ClientInfo, topic string) bool
	ProcessUnSub(clientInfo *ClientInfo, topic string)
	SetAuthFunc(f GetAuthFunc)
}

var providers = make(map[string]Authenticator)

type TopicAclManger struct {
	p Authenticator
}

func (this *TopicAclManger) CheckPub(clientInfo *ClientInfo, topic string) bool {
	return this.p.CheckPub(clientInfo, topic)
}

func (this *TopicAclManger) CheckSub(clientInfo *ClientInfo, topic string) bool {
	return this.p.CheckSub(clientInfo, topic)
}

func (this *TopicAclManger) ProcessUnSub(clientInfo *ClientInfo, topic string) {
	this.p.ProcessUnSub(clientInfo, topic)
	return
}

func (this *TopicAclManger) SetAuthFunc(f GetAuthFunc) {
	this.p.SetAuthFunc(f)
}

func NewTopicAclManger(providerName string, f GetAuthFunc) (*TopicAclManger, error) {
	if len(providerName) == 0 {
		return nil, errors.New("providerName or f invalid !")
	}
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

func log(subPub, topic string, clientInfo *ClientInfo) {
	if logger == nil {
		fmt.Println("Logger == nil ")
		return
	}
	logFields := []zapcore.Field{}
	logFields = append(logFields, zap.String("topic", topic))
	if clientInfo != nil {
		logFields = append(logFields, zap.String("user-name", clientInfo.UserName))
		logFields = append(logFields, zap.String("user-id", clientInfo.UserId))
		if clientInfo.SdkInfo != nil {
			//sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64
			logFields = append(logFields, zap.String("sdk-lang", clientInfo.SdkInfo.SdkLang))
			logFields = append(logFields, zap.String("sdk-version", clientInfo.SdkInfo.SdkVersion))
			logFields = append(logFields, zap.String("sdk-arch", clientInfo.SdkInfo.SdkArch))
			logFields = append(logFields, zap.String("sdk-os", clientInfo.SdkInfo.SdkOs))
		}
	}
	logger.Info(subPub, logFields...)
}
