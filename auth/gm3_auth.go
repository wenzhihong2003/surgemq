package auth

import (
	"github.com/wenzhihong2003/surgemq/slog"
)

var _ Authenticator = (*gm3Authenticator)(nil)

type gm3Authenticator struct {
	authFunc AuthFunc
}

type AuthFunc func(token string) (bool, *ClientInfo)

func (this *gm3Authenticator) Authenticate(token, userName string) (bool, *ClientInfo) {
	if len(token) == 0 {
		return false, nil
	}

	// 如果没有设置authFunc认为验证通过
	if this.authFunc == nil {
		return true, nil
	}

	// 通过token获取信息
	isPass, info := this.authFunc(token)
	slog.Infof("surgemq 调用auth认证token结果, token=%s, isPass=%t, clientInfo=%s", token, isPass, info.String())
	return isPass, info
}

func (this *gm3Authenticator) SetAuthFunc(f AuthFunc) {
	this.authFunc = f
}

func init() {
	Register("gm3Auth", new(gm3Authenticator))
}
