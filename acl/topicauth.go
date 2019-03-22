package acl

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/wenzhihong2003/message"
	"github.com/wenzhihong2003/surgemq/slog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TopicAlwaysVerifyType = "topicAlwaysVerify"
	TopicNumAuthType      = "topicNumAuth"
	TopicSetAuthType      = "topicSetAuth"
	userTopicKeyFmt       = "%s:%s"
)

type GetAuthFunc func(clientInfo *ClientInfo, topic string) interface{}

type ClientInfo struct {
	GmToken        string // mqtt password
	GmUserName     string
	GmUserId       string
	GmSdkInfo      map[string]string // from mqtt userName
	ConnectMessage *message.ConnectMessage
	SubTopicLimit  int
	LocalAddr      string
	RemoteAddr     string
}

func (this *ClientInfo) String() string {
	if this == nil {
		return "AuthClientInfo nil"
	}

	buf := bytes.Buffer{}

	buf.WriteString("AuthClientInfo GmToken=")
	buf.WriteString(this.GmToken)

	buf.WriteString(" GmUserName=")
	buf.WriteString(this.GmUserName)

	buf.WriteString(" SubTopicLimit=")
	buf.WriteString(strconv.Itoa(this.SubTopicLimit))

	buf.WriteString(" LocalAddr=")
	buf.WriteString(this.LocalAddr)

	buf.WriteString(" RemoteAddr=")
	buf.WriteString(this.RemoteAddr)

	buf.WriteString(" GmSdkInfo[")
	for k, v := range this.GmSdkInfo {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(v)
		buf.WriteString(";")
	}
	buf.WriteString("]")

	return buf.String()
}

// sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64
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

const (
	userId     = "userid"
	sdkArch    = "sdk-arch"
	sdkOs      = "sdk-os"
	sdkLang    = "sdk-lang"
	sdkVersion = "sdk-version"
	unKnown    = "unknown"
)

func operLog(subPub, topic string, clientInfo *ClientInfo) {
	if slog.OperationLogger == nil {
		return
	}
	logMap := make(map[string]string)
	logMap[userId] = unKnown
	logMap[sdkArch] = unKnown
	logMap[sdkOs] = unKnown
	logMap[sdkLang] = unKnown
	logMap[sdkVersion] = unKnown
	if clientInfo != nil {
		// logFields = append(logFields, zap.String("user-name", clientInfo.GmUserName))
		logMap[userId] = clientInfo.GmUserId
		for k, v := range clientInfo.GmSdkInfo {
			logMap[k] = v
		}
	}
	logFields := []zapcore.Field{zap.String("type", subPub), zap.String("topic", topic)}
	for k, v := range logMap {
		logFields = append(logFields, zap.String(k, v))
	}
	slog.OperationLogger.Info("acl", logFields...)
}
