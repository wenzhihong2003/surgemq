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

package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wenzhihong2003/message"
	"github.com/wenzhihong2003/surgemq/acl"
	"github.com/wenzhihong2003/surgemq/auth"
	"github.com/wenzhihong2003/surgemq/sessions"
	"github.com/wenzhihong2003/surgemq/slog"
	"github.com/wenzhihong2003/surgemq/topics"
)

var (
	ErrInvalidConnectionType  error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
)

const (
	DefaultKeepAlive        = 300
	DefaultConnectTimeout   = 2
	DefaultAckTimeout       = 20
	DefaultTimeoutRetries   = 3
	DefaultSessionsProvider = "mem"
	DefaultAuthenticator    = "mockSuccess"
	DefaultTopicsProvider   = "mem"
)

// Server is a library implementation of the MQTT server that, as best it can, complies
// with the MQTT 3.1 and 3.1.1 specs.
type Server struct {
	// The number of seconds to keep the connection live if there's no data.
	// If not set then default to 5 mins.
	KeepAlive int

	// The number of seconds to wait for the CONNECT message before disconnecting.
	// If not set then default to 2 seconds.
	ConnectTimeout int

	// The number of seconds to wait for any ACK messages before failing.
	// If not set then default to 20 seconds.
	AckTimeout int

	// The number of times to retry sending a packet if ACK is not received.
	// If no set then default to 3 retries.
	TimeoutRetries int

	// Authenticator is the authenticator used to check username and password sent
	// in the CONNECT message. If not set then default to "mockSuccess".
	Authenticator string

	// SessionsProvider is the session store that keeps all the Session objects.
	// This is the store to check if CleanSession is set to 0 in the CONNECT message.
	// If not set then default to "mem".
	SessionsProvider string

	// TopicsProvider is the topic store that keeps all the subscription topics.
	// If not set then default to "mem".
	TopicsProvider string

	// authMgr is the authentication manager that we are going to use for authenticating
	// incoming connections
	authMgr *auth.Manager

	// sessMgr is the sessions manager for keeping track of the sessions
	sessMgr *sessions.Manager

	// topicsMgr is the topics manager for keeping track of subscriptions
	topicsMgr *topics.Manager

	// The quit channel for the server. If the server detects that this channel
	// is closed, then it's a signal for it to shutdown as well.
	quit chan struct{}

	ln net.Listener

	// A list of services created by the server. We keep track of them so we can
	// gracefully shut them down if they are still alive when the server goes down.
	// 这个变量没有用了
	svcs []*service
	// 用于保存当前的 services
	svcmap           map[uint64]*service
	serviceCountStat serviceCountStat

	// Mutex for updating svcs, svcmap
	mu sync.Mutex

	// A indicator on whether this server is running
	running int32

	// A indicator on whether this server has already checked configuration
	configOnce sync.Once

	TopicAclFunc acl.GetAuthFunc
	AuthFunc     auth.AuthFunc
	AclProvider  string
	aclManger    *acl.TopicAclManger

	pubStat stat
}

type connInfo struct {
	GmUserName      string            `json:"gmUserName"`
	GmUserId        string            `json:"gmUserId"`
	GmSdkInfo       map[string]string `json:"gmSdkInfo"`
	SubTopicLimit   int               `json:"subTopicLimit"`
	SubedTopicCount int32             `json:"subedTopicCount"` // 已订阅的主题数
	LocalAddr       string            `json:"localAddr"`
	RemoteAddr      string            `json:"remoteAddr"`
	InBytes         int64             `json:"inBytes"`
	InMsgs          int64             `json:"inMsgs"`
	OutBytes        int64             `json:"outBytes"`
	OutMsgs         int64             `json:"outMsgs"`
	DiscardPubMsg   int64             `json:"discardPubMsg"`
}

// 得到每个连接的信息
func (this *Server) GetServiceConnInfo() []connInfo {
	if this == nil {
		return []connInfo{}
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	result := make([]connInfo, 0, len(this.svcmap))
	for _, svc := range this.svcmap {
		clientInfo := svc.clientInfo
		if clientInfo != nil {
			result = append(result, connInfo{
				GmUserName:      clientInfo.GmUserName,
				GmUserId:        clientInfo.GmUserId,
				GmSdkInfo:       clientInfo.GmSdkInfo,
				SubTopicLimit:   clientInfo.SubTopicLimit,
				SubedTopicCount: atomic.LoadInt32(&svc.subedTopicCount),
				LocalAddr:       clientInfo.LocalAddr,
				RemoteAddr:      clientInfo.RemoteAddr,
				InBytes:         svc.inStat.bytes,
				InMsgs:          svc.inStat.msgs,
				OutBytes:        svc.outStat.bytes,
				OutMsgs:         svc.outStat.msgs,
				DiscardPubMsg:   svc.discardPubMsg,
			})
		}
	}
	return result
}

func (this *Server) GetServiceCountStatInfoData() map[string]interface{} {
	if this != nil {
		running := false
		if atomic.LoadInt32(&this.running) == 1 {
			running = true
		}
		return map[string]interface{}{
			"running":  running,
			"online":   this.serviceCountStat.online,
			"downline": this.serviceCountStat.downline,
			"total":    this.serviceCountStat.total,
			"pubMsg":   atomic.LoadInt64(&this.pubStat.msgs),
			"pubBytes": atomic.LoadInt64(&this.pubStat.bytes),
		}
	} else {
		return map[string]interface{}{}
	}
}

func (this *Server) GetServiceCountStatInfo() string {
	if this != nil {
		return "running=" + strconv.Itoa(int(this.running)) +
			" " + this.serviceCountStat.String() +
			"  发布消息 " + strconv.FormatInt(atomic.LoadInt64(&this.pubStat.msgs), 10) +
			"  个 字节数 " + strconv.FormatInt(atomic.LoadInt64(&this.pubStat.bytes), 10)

	} else {
		return ""
	}
}

// 启动server, 在退出时不关闭Server, 要求在应用层进行关闭
func (this *Server) ListenAndServeWithoutCloseServer(uri string) error {
	return this._listenAndServe(uri, false)
}

func (this *Server) ListenAndServe(uri string) error {
	return this._listenAndServe(uri, true)
}

// ListenAndServe listents to connections on the URI requested, and handles any
// incoming MQTT client sessions. It should not return until Close() is called
// or if there's some critical error that stops the server from running. The URI
// supplied should be of the form "protocol://host:port" that can be parsed by
// url.Parse(). For example, an URI could be "tcp://0.0.0.0:1883".
func (this *Server) _listenAndServe(uri string, closeServerInReturn bool) error {
	defer atomic.CompareAndSwapInt32(&this.running, 1, 0)

	if !atomic.CompareAndSwapInt32(&this.running, 0, 1) {
		return fmt.Errorf("server/ListenAndServe: Server is already running")
	}

	this.quit = make(chan struct{})

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	this.ln, err = net.Listen(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	defer this.ln.Close()
	if closeServerInReturn {
		defer this.Close()
	}
	slog.Infof("server/ListenAndServe: mqttserver 已准备好接受客户端请求...")

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := this.ln.Accept()

		if err != nil {
			// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
			select {
			case <-this.quit:
				return nil

			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				slog.Errorf("server/ListenAndServe: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go this.handleConnection(conn)
	}
}

// 1. tls读取服务器证书，没有根证书的类型
func (this *Server) TlsListenAndServeWithByte(uri string, certPEMBlock, keyPEMBlock []byte) error {
	defer atomic.CompareAndSwapInt32(&this.running, 1, 0)

	if !atomic.CompareAndSwapInt32(&this.running, 0, 1) {
		return fmt.Errorf("server/ListenAndServe: Server is already running")
	}

	this.quit = make(chan struct{})

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	// server 端加入 tls
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		fmt.Println(err)
		return err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
	}
	this.ln, err = tls.Listen("tcp", u.Host, config)
	if err != nil {
		log.Println(err)
		return err
	}
	defer this.ln.Close()

	slog.Infof("server/ListenAndServe: server is ready...")

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := this.ln.Accept()

		if err != nil {
			// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
			select {
			case <-this.quit:
				return nil

			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				slog.Errorf("server/ListenAndServe: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go this.handleConnection(conn)
	}
}

// 2.tls 服务器证书和根证书并用的类型
func (this *Server) TlsListenAndServeWithFile(uri, ca, crt, key string) error {
	defer atomic.CompareAndSwapInt32(&this.running, 1, 0)

	if !atomic.CompareAndSwapInt32(&this.running, 0, 1) {
		return fmt.Errorf("server/ListenAndServe: Server is already running")
	}

	this.quit = make(chan struct{})

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	// read crt and key
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(ca)
	if err != nil {
		fmt.Println(err)
		return err
	}
	certpool.AppendCertsFromPEM(pemCerts)
	// server 端加入 tls
	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		fmt.Println(err)
		return err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certpool,
	}
	this.ln, err = tls.Listen("tcp", u.Host, config)
	if err != nil {
		log.Println(err)
		return err
	}
	defer this.ln.Close()

	slog.Infof("server/ListenAndServe: server is ready...")

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := this.ln.Accept()

		if err != nil {
			// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
			select {
			case <-this.quit:
				return nil

			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				slog.Errorf("server/ListenAndServe: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go this.handleConnection(conn)
	}
}

// Publish sends a single MQTT PUBLISH message to the server. On completion, the
// supplied OnCompleteFunc is called. For QOS 0 messages, onComplete is called
// immediately after the message is sent to the outgoing buffer. For QOS 1 messages,
// onComplete is called when PUBACK is received. For QOS 2 messages, onComplete is
// called after the PUBCOMP message is received.
func (this *Server) Publish(msg *message.PublishMessage, onComplete OnCompleteFunc) error {
	if err := this.checkConfiguration(); err != nil {
		return err
	}

	if msg.Retain() {
		if err := this.topicsMgr.Retain(msg); err != nil {
			slog.Errorf("Error retaining message: %v", err)
		}
	}

	svcLen := len(this.svcmap)
	subs := make([]interface{}, 0, svcLen)
	qoss := make([]byte, 0, svcLen)

	if err := this.topicsMgr.Subscribers(msg.Topic(), msg.QoS(), &subs, &qoss); err != nil {
		return err
	}

	msg.SetRetain(false)

	slog.Debugf("(server) Publishing to topic %q and %d subscribers", string(msg.Topic()), len(subs))
	for _, s := range subs {
		if s != nil {
			fn, ok := s.(*OnPublishFunc)
			if !ok {
				slog.Errorf("Invalid onPublish Function")
			} else {
				(*fn)(msg)
				this.pubStat.increment(int64(msg.Len()))
			}
		}
	}

	return nil
}

// Close terminates the server by shutting down all the client connections and closing
// the listener. It will, as best it can, clean up after itself.
func (this *Server) Close() error {
	// By closing the quit channel, we are telling the server to stop accepting new
	// connection.
	close(this.quit)

	// We then close the net.Listener, which will force Accept() to return if it's
	// blocked waiting for new connections.
	this.ln.Close()

	slog.Infof("at close server, server's serviceCountStatus %s", this.serviceCountStat.String())

	// this.mu.Lock()
	// for _, svc := range this.svcs {
	// 	if !svc.isDone() {
	// 		slog.Debugf("Stopping service %d", svc.id)
	// 		svc.stop()
	// 	}
	// }
	for _, svc := range this.svcmap {
		if !svc.isDone() {
			slog.Debugf("Stopping service %d", svc.id)
			svc.stop()
		}
	}
	// this.mu.Unlock()

	if this.sessMgr != nil {
		this.sessMgr.Close()
	}

	if this.topicsMgr != nil {
		this.topicsMgr.Close()
	}

	return nil
}

// HandleConnection is for the broker to handle an incoming connection from a client
func (this *Server) handleConnection(c io.Closer) (svc *service, err error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	err = this.checkConfiguration()
	if err != nil {
		return nil, err
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	// To establish a connection, we must
	// 1. Read and decode the message.ConnectMessage from the wire
	// 2. If no decoding errors, then authenticate using username and password.
	//    Otherwise, write out to the wire message.ConnackMessage with
	//    appropriate error.
	// 3. If authentication is successful, then either create a new session or
	//    retrieve existing session
	// 4. Write out to the wire a successful message.ConnackMessage message

	// Read the CONNECT message from the wire, if error, then check to see if it's
	// a CONNACK error. If it's CONNACK error, send the proper CONNACK error back
	// to client. Exit regardless of error type.

	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(this.ConnectTimeout)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(conn)
	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			// slog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return nil, err
	}

	// Authenticate the user, if error, return error and exit
	// password存着token
	verify, clientInfo := this.authMgr.Authenticate(string(req.Password()), string(req.Username()))
	if !verify {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return nil, err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	aclClientInfo := getClientInfo(clientInfo, req)
	svc = &service{
		id:     atomic.AddUint64(&gsvcid, 1),
		client: false,

		keepAlive:      int(req.KeepAlive()),
		connectTimeout: this.ConnectTimeout,
		ackTimeout:     this.AckTimeout,
		timeoutRetries: this.TimeoutRetries,

		conn:       conn,
		sessMgr:    this.sessMgr,
		topicsMgr:  this.topicsMgr,
		aclManger:  this.aclManger,
		clientInfo: aclClientInfo,

		ownerServer: this,

		done: make(chan struct{}, 1),

		pubMsgChan: make(chan *message.PublishMessage, gPubMsgChanSize),
	}

	if aclClientInfo != nil && conn != nil {
		aclClientInfo.RemoteAddr = addr2str(conn.RemoteAddr())
		aclClientInfo.LocalAddr = addr2str(conn.LocalAddr())
	}

	// 设置一下clientid
	if aclClientInfo != nil && len(req.ClientId()) == 0 && len(aclClientInfo.GmUserId) > 0 {
		req.SetClientId([]byte(fmt.Sprintf("gmuserid_%s", aclClientInfo.GmUserId)))
		req.SetCleanSession(true)
	}

	err = this.getSession(svc, req, resp)
	if err != nil {
		return nil, err
	}

	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(c, resp); err != nil {
		return nil, err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	if err := svc.start(); err != nil {
		svc.stop()
		return nil, err
	}

	this.mu.Lock()
	if len(this.svcmap) == 0 {
		this.svcmap = make(map[uint64]*service, 2048)
	}
	this.svcmap[svc.id] = svc
	// if len(this.svcs) == 0 {
	// 	this.svcs = make([]*service, 0, 2048)
	// }
	// this.svcs = append(this.svcs, svc)
	this.mu.Unlock()

	slog.Infof("(%s) server/handleConnection: Connection established.", svc.cid())

	return svc, nil
}

func (this *Server) removeService(srv *service) {
	this.mu.Lock()
	if this != nil && srv != nil {
		delete(this.svcmap, srv.id)
	}
	this.mu.Unlock()
}

/*
Password: bearer 1bcf468df513a81bf9fdf698694a327d5fba12b7
Username: sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64
*/
func getClientInfo(clientInfo *auth.ClientInfo, req *message.ConnectMessage) *acl.ClientInfo {
	// todo 兼容旧版本没有token的
	if clientInfo == nil {
		return nil
	}

	sdkInfoStr := string(req.Username())
	if len(sdkInfoStr) == 0 {
		return &acl.ClientInfo{GmToken: clientInfo.Token}
	}

	// 解析sdkinfo
	sdkInfo := map[string]string{}
	for _, str := range strings.Split(sdkInfoStr, "|") {
		if kv := strings.Split(str, "="); len(kv) == 2 {
			sdkInfo[kv[0]] = kv[1]
		}
	}
	return &acl.ClientInfo{
		GmToken:        clientInfo.Token,
		GmUserName:     clientInfo.UserName,
		GmUserId:       clientInfo.UserId,
		GmSdkInfo:      sdkInfo,
		ConnectMessage: req,
		SubTopicLimit:  clientInfo.SubTopicLimit,
	}

}

func (this *Server) checkConfiguration() error {
	var err error

	this.configOnce.Do(func() {
		if this.KeepAlive == 0 {
			this.KeepAlive = DefaultKeepAlive
		}

		if this.ConnectTimeout == 0 {
			this.ConnectTimeout = DefaultConnectTimeout
		}

		if this.AckTimeout == 0 {
			this.AckTimeout = DefaultAckTimeout
		}

		if this.TimeoutRetries == 0 {
			this.TimeoutRetries = DefaultTimeoutRetries
		}

		if this.Authenticator == "" {
			this.Authenticator = "mockSuccess"
		}

		this.authMgr, err = auth.NewManager(this.Authenticator, this.AuthFunc)
		if err != nil {
			slog.Errorf("surgemq checkConfiguration has error, err=%v", err)
			return
		}

		if this.SessionsProvider == "" {
			this.SessionsProvider = "mem"
		}

		this.sessMgr, err = sessions.NewManager(this.SessionsProvider)
		if err != nil {
			return
		}

		if this.TopicsProvider == "" {
			this.TopicsProvider = "mem"
		}

		this.topicsMgr, err = topics.NewManager(this.TopicsProvider)
		if err != nil {
			return
		}

		if this.AclProvider == "" {
			this.AclProvider = acl.TopicAlwaysVerifyType
		}

		this.aclManger, err = acl.NewTopicAclManger(this.AclProvider, this.TopicAclFunc)

		return
	})

	return err
}

func (this *Server) getSession(svc *service, req *message.ConnectMessage, resp *message.ConnackMessage) error {
	// If CleanSession is set to 0, the server MUST resume communications with the
	// client based on state from the current session, as identified by the client
	// identifier. If there is no session associated with the client identifier the
	// server must create a new session.
	//
	// If CleanSession is set to 1, the client and server must discard any previous
	// session and start a new one. This session lasts as long as the network c
	// onnection. State data associated with this session must not be reused in any
	// subsequent session.

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient%d", svc.id)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if svc.sess, err = this.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := svc.sess.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if svc.sess == nil {
		if svc.sess, err = this.sessMgr.New(cid); err != nil {
			return err
		}

		resp.SetSessionPresent(false)

		if err := svc.sess.Init(req); err != nil {
			return err
		}
	}

	return nil
}
