// Copyright (c) 2014 The SurgeMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrAuthFailure          = errors.New("auth: Authentication failure")
	ErrAuthProviderNotFound = errors.New("auth: Authentication provider not found")

	providers = make(map[string]Authenticator)
)

type ClientInfo struct {
	Token         string
	UserName      string
	UserId        string
	SubTopicLimit int
}

func (this *ClientInfo) String() string {
	if this == nil {
		return "auth.ClientInfo: nil"
	}
	buf := bytes.Buffer{}
	buf.WriteString("auth.ClientInfo: Token=")
	buf.WriteString(this.Token)
	buf.WriteString(" UserName=")
	buf.WriteString(this.UserName)
	buf.WriteString(" UserId=")
	buf.WriteString(this.UserId)
	buf.WriteString(" SubTopicLimit=")
	buf.WriteString(strconv.Itoa(this.SubTopicLimit))
	return buf.String()
}

type Authenticator interface {
	Authenticate(token, userName string) (bool, *ClientInfo)
	SetAuthFunc(f AuthFunc)
}

func Register(name string, provider Authenticator) {
	if provider == nil {
		panic("auth: Register provide is nil")
	}

	if _, dup := providers[name]; dup {
		panic("auth: Register called twice for provider " + name)
	}

	providers[name] = provider
}

func Unregister(name string) {
	delete(providers, name)
}

type Manager struct {
	p Authenticator
}

func NewManager(providerName string, authFunc AuthFunc) (*Manager, error) {
	p, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provider %q", providerName)
	}
	p.SetAuthFunc(authFunc)

	return &Manager{p: p}, nil
}

func (this *Manager) Authenticate(token, userName string) (bool, *ClientInfo) {
	return this.p.Authenticate(token, userName)
}

func (this *Manager) SetAuthFunc(f AuthFunc) {
	this.p.SetAuthFunc(f)
}
