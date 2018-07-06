package benchmark

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	MQTT "github.com/liaoliaopro/paho.mqtt.golang"
	"go.uber.org/zap"

	"crypto/x509"

	"github.com/fangwendong/surgemq/acl"
	"github.com/fangwendong/surgemq/service"
)

func Test1(t *testing.T) {
	var f MQTT.MessageHandler = func(MQTT.Client, MQTT.Message) {
		fmt.Println("rece")
	}

	connOpts := &MQTT.ClientOptions{
		ClientID:             "ds-live",
		CleanSession:         true,
		Username:             "",
		Password:             "",
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            30 * time.Second,
		AutoReconnect:        true,
		PingTimeout:          10 * time.Second,
		ConnectTimeout:       30 * time.Second,
		TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
		OnConnectionLost:     func(c MQTT.Client, err error) { fmt.Println("mqtt disconnected.", zap.Error(err)) },
	}
	connOpts.AddBroker("tcp://testserver:8011")

	mc := MQTT.NewClient(connOpts)
	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("mqtt connection failed.", zap.Error(token.Error()))
		return
	}

	defer mc.Disconnect(12)
	for {
		if token := mc.Publish("ds", 0, false, "xx"); token.Wait() && token.Error() != nil {
			fmt.Println("publish failed.", zap.Error(token.Error()))

		}
	}

	if token := mc.Subscribe("dsx", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println("publish failed.", zap.Error(token.Error()))

	} else {
		t := token.(*MQTT.SubscribeToken)
		fmt.Println(t.Result())
	}

	if token := mc.Subscribe("ds", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println("publish failed.", zap.Error(token.Error()))

	} else {
		t := token.(*MQTT.SubscribeToken)
		fmt.Println(t.Result())
	}

	if token := mc.Unsubscribe("ds"); token.Wait() && token.Error() != nil {
		fmt.Println("publish failed.", zap.Error(token.Error()))

	}
	if token := mc.Subscribe("dsx", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println("publish failed.", zap.Error(token.Error()))

	} else {
		t := token.(*MQTT.SubscribeToken)
		fmt.Println(t.Result())
	}

}

func Test2(t *testing.T) {
	mqttServer := &service.Server{
		KeepAlive:        300,           // seconds
		ConnectTimeout:   2,             // seconds
		SessionsProvider: "mem",         // keeps sessions in memory
		Authenticator:    "mockSuccess", // always succeed
		TopicsProvider:   "mem",         // keeps topic subscriptions in memory
		AclProvider:      acl.TopicNumAuthType,
		TopicAclFunc: func(info *acl.ClientInfo, topic string) interface{} {
			return 1
		},
	}

	if err := mqttServer.ListenAndServe("tcp://127.0.0.1:8081"); err != nil {
		fmt.Println("mqtt error", zap.Error(err))
	}
}

func TestTggw(t *testing.T) {
	mqttServer := &service.Server{
		KeepAlive:        300,           // seconds
		ConnectTimeout:   2,             // seconds
		SessionsProvider: "mem",         // keeps sessions in memory
		Authenticator:    "mockSuccess", // always succeed
		TopicsProvider:   "mem",         // keeps topic subscriptions in memory
		AclProvider:      acl.TopicAlwaysVerifyType,
	}

	if err := mqttServer.ListenAndServe("tcp://127.0.0.1:8080"); err != nil {
		fmt.Println("mqtt error", zap.Error(err))
	}
}

func Test3(t *testing.T) {
	mqttServer := &service.Server{
		KeepAlive:        300,           // seconds
		ConnectTimeout:   2,             // seconds
		SessionsProvider: "mem",         // keeps sessions in memory
		Authenticator:    "mockSuccess", // always succeed
		TopicsProvider:   "mem",         // keeps topic subscriptions in memory
		AclProvider:      acl.TopicSetAuthType,
		TopicAclFunc: func(info *acl.ClientInfo, topic string) interface{} {
			return true
		},
	}

	if err := mqttServer.ListenAndServe("tcp://127.0.0.1:8080"); err != nil {
		fmt.Println("mqtt error", zap.Error(err))
	}
}

func Test4(t *testing.T) {
	var f MQTT.MessageHandler = func(c MQTT.Client, m MQTT.Message) {
		// fmt.Println("rece:", m.Topic())
	}
	var ch chan int
	connOpts := &MQTT.ClientOptions{
		ClientID:             "ds-live",
		CleanSession:         true,
		Username:             "sdk-lang=python3.6|sdk-version=3.0.0.96|sdk-arch=64|sdk-os=win-amd64",
		Password:             "faa4e334845d00bbb2e03ee524eda540ce2672dd",
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            30 * time.Second,
		AutoReconnect:        true,
		PingTimeout:          10 * time.Second,
		ConnectTimeout:       30 * time.Second,
		TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
		OnConnectionLost:     func(c MQTT.Client, err error) { fmt.Println("mqtt disconnected.", zap.Error(err)) },
	}
	connOpts.AddBroker("tcp://ds-live:7021")

	mc := MQTT.NewClient(connOpts)
	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("mqtt connection failed.", zap.Error(token.Error()))
		return
	}
	// defer mc.Disconnect(12)

	ts := make([]string, 1000)

	for i := 0; i <= 1000; i++ {
		if 0 < i && i < 10 {
			ts[i] = fmt.Sprintf("pb/data.api.Tick/SZSE/00000%d", i)
			continue
		}

		if i < 100 {
			ts[i] = fmt.Sprintf("pb/data.api.Tick/SZSE/0000%d", i)
			continue
		}

		if i > 100 && i < 1000 {
			ts[i] = fmt.Sprintf("pb/data.api.Tick/SZSE/000%d", i)
			continue
		}
	}
	for j := 1; j <= 900; j++ {
		go func(i int) {
			if token := mc.Subscribe(ts[i], 0, f); token.Wait() && token.Error() != nil {
				fmt.Println("sub failed.", zap.Error(token.Error()))

			} else {
				t := token.(*MQTT.SubscribeToken)
				fmt.Println(t.Result())
			}
		}(j)

		// if token := mc.Subscribe(ts[j], 0, f); token.Wait() && token.Error() != nil {
		// 	fmt.Println("sub failed.", zap.Error(token.Error()))
		//
		// } else {
		// 	t := token.(*MQTT.SubscribeToken)
		// 	fmt.Println(t.Result())
		// }

	}

	<-ch
}

func TestTlsServer(t *testing.T) {
	mqttServer := &service.Server{
		KeepAlive:        300,           // seconds
		ConnectTimeout:   2,             // seconds
		SessionsProvider: "mem",         // keeps sessions in memory
		Authenticator:    "mockSuccess", // always succeed
		TopicsProvider:   "mem",         // keeps topic subscriptions in memory
		AclProvider:      acl.TopicAlwaysVerifyType,
	}

	if err := mqttServer.TlsListenAndServeWithByte("tcp://127.0.0.1:8080",
		[]byte(`-----BEGIN CERTIFICATE-----
MIIC9zCCAd+gAwIBAgIJALRgm2/ZL3NDMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMMB215cXVhbnQwHhcNMTgwNTA5MDI1NzAwWhcNMzIwMTE2MDI1NzAwWjASMRAw
DgYDVQQDDAdteXF1YW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
9cQHRYOM3VRitCfvuYfpbntOFljFlRTWvA2ECfEhQaV9ydmM+zL5rlPMo1aE5U0g
r8IU8uRPcm3S5FbxHIba/lCiE+SkUKf7JAYminKYFtnBxterr5ovolK2WBNMy83x
q/GtiFlLy9A03BPPbRu6DS6yXZ49c3gskNyRjUR9KQ4A+fuEUXQLTl0mUOn2jniq
V3rebnz2ooQbjTz02l/En3ITIxDmiIcGZBdYMdw+T2eIhIx0uMb84KGnIGQddV8X
zoin9c60NzB9A18QfaaiGs3ISJTEizlnJpc/fD3AEfidjrxfLbNVLoCRix8a7Ct/
4hVCluVKKtuGeBVKXqtopQIDAQABo1AwTjAdBgNVHQ4EFgQU1AX4BBN2xuUCJ5OU
GKHcwf9mAy0wHwYDVR0jBBgwFoAU1AX4BBN2xuUCJ5OUGKHcwf9mAy0wDAYDVR0T
BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAWL4XC2PR/hOD5fxdOIOAbS1B9S6Z
9ft1aznd8/fvaFwkVXXgZuzXCyJWdG7WiVwxk4JeUJYO7CShEFKf1W4b0HLVSFHf
Ns9bpEnFCGoIm+nEWHxfJIiKcKw48vOLBE7DLZZTyJFzuowSjA9uHXWPTSc3YW8Q
vmUEzvRA54+uezBZWK0YcxE+LqAfTaWzFjHbRaztxvf27XkQL+tNbBq8c3FCTEF2
ZQa0pP7IUM+wy3Lw7r8guZxzzhI1zU2MazFsb0R2z/mGPMdGDVF+kFZk6UYFm8Pm
aPM590xia5BCDOWAMoz/qzalR4KXrndxhfwSxtceMojAbok5VxOvuOdgOg==
-----END CERTIFICATE-----`),
		[]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA9cQHRYOM3VRitCfvuYfpbntOFljFlRTWvA2ECfEhQaV9ydmM
+zL5rlPMo1aE5U0gr8IU8uRPcm3S5FbxHIba/lCiE+SkUKf7JAYminKYFtnBxter
r5ovolK2WBNMy83xq/GtiFlLy9A03BPPbRu6DS6yXZ49c3gskNyRjUR9KQ4A+fuE
UXQLTl0mUOn2jniqV3rebnz2ooQbjTz02l/En3ITIxDmiIcGZBdYMdw+T2eIhIx0
uMb84KGnIGQddV8Xzoin9c60NzB9A18QfaaiGs3ISJTEizlnJpc/fD3AEfidjrxf
LbNVLoCRix8a7Ct/4hVCluVKKtuGeBVKXqtopQIDAQABAoIBACxC+LhJP+Zgm0Is
6xyz4JOYO3EE3djCnwXRPVV5Cu9CucvPlXdsL9F4oCNOorKVWGyu/IzeF3zZUKD1
E4l6tBgXN4lzpTAtgKp1PR20E7YR07dXAkGm+Cs40EQ+fZc66Op5pTIoOuFjBxIX
s0TIOJLFNTRtGk4gd45DWydhDVWGKMSLyB8iaHJ7kHOEGjML5YcL6hMkD1n6TCPa
0RTl5VaDM7FzKB46nn6Y4KAIC34IRU5vO+uqAEBaB9C5wkLO3ZaWeW8Cgy8kNNFy
EBlUqUaCgqurZcZg4YPordRDQW9VQMOdGuv4js+5HD93K2adpI3towaBjhzrdoEo
TlFo6WUCgYEA/uSJagFpI/Djegy8neXGsj1N8XdqP9T5kgv/xBWlaIPzr0SBQlxW
U8anWATITql2iZr1F768LTk4dfIYSIKcQshw0YnhP7qb7FY+ceFvOCiUT9A+uQ1w
QIOJr8ydB8PffrqWL68il1qk19+4Dvnpsn0cY6BD9szpUCdauIqlorcCgYEA9tVX
dCbocsl3cLonzM1vWTqGyD5oGDWMqP6azadyoLlLSRGfB0jppjQ20pZBxrJq/cTy
Nl8uffbJul8g1/6Dk8WyKD7Fei1A9BCN4QO+w6ZcD69GHX+X6xUXdRMM2SFLaoGD
7+vN9CdHBftn2iFjHRGjJ8sWQvATq3JDxDgHA4MCgYA+blOJ91Z9Sx8sYbpBImqM
dZ+FqS4I/G00bGP07yhYdRlWsHzIeD1cv6d1U5aMTc2O3rlxW3JT0VQW73krKXKE
mPupFxBov5g5RtZ8pi7LnoTVF7iFMtlvs8ghmwhLQpqXO7RVcZwTXkxJ4639XRD0
ethdPn/nD0GGNF1wHeV7+QKBgFEcikCBCKGv0rAYfDuwxoZr1R64YzyRXEesYvJx
tBlcyoCYacnbC+yx+9H3zmWc+8uojG+Rl5WNI307BW/1EwfcT08qUXp0pIOPbRAk
SuvAH0CIOGI5K5L0u2CdgftYFZBKPzD4LBWvUoeEtfvYPNmwkgzhj88vVUdhpSM1
xhhBAoGBAJ+Wq25obHvrnBHNJzMsbcAEk8MHVW1Cb+k7Sa1qKGUdwHEnUyWVw3FI
V6gSIBQvV1gUP1o/53nFDqmXbWzg3zfUx8PAea4RU0hUSbLOZcmJnwzARzIL7ap/
V3VTQ3Ek0gD3+VAXLNiiXk9XUFVoFFVrwWUeBtxSw3B7Fmpi36ME
-----END RSA PRIVATE KEY-----`),
	); err != nil {
		fmt.Println("mqtt error", zap.Error(err))
	}
}

func NewTLSConfig(ca, crt, key string) *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	//pemCerts, err := ioutil.ReadFile(ca)

	certpool.AppendCertsFromPEM([]byte(`-----BEGIN CERTIFICATE-----
MIIC9zCCAd+gAwIBAgIJALRgm2/ZL3NDMA0GCSqGSIb3DQEBCwUAMBIxEDAOBgNV
BAMMB215cXVhbnQwHhcNMTgwNTA5MDI1NzAwWhcNMzIwMTE2MDI1NzAwWjASMRAw
DgYDVQQDDAdteXF1YW50MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
9cQHRYOM3VRitCfvuYfpbntOFljFlRTWvA2ECfEhQaV9ydmM+zL5rlPMo1aE5U0g
r8IU8uRPcm3S5FbxHIba/lCiE+SkUKf7JAYminKYFtnBxterr5ovolK2WBNMy83x
q/GtiFlLy9A03BPPbRu6DS6yXZ49c3gskNyRjUR9KQ4A+fuEUXQLTl0mUOn2jniq
V3rebnz2ooQbjTz02l/En3ITIxDmiIcGZBdYMdw+T2eIhIx0uMb84KGnIGQddV8X
zoin9c60NzB9A18QfaaiGs3ISJTEizlnJpc/fD3AEfidjrxfLbNVLoCRix8a7Ct/
4hVCluVKKtuGeBVKXqtopQIDAQABo1AwTjAdBgNVHQ4EFgQU1AX4BBN2xuUCJ5OU
GKHcwf9mAy0wHwYDVR0jBBgwFoAU1AX4BBN2xuUCJ5OUGKHcwf9mAy0wDAYDVR0T
BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAWL4XC2PR/hOD5fxdOIOAbS1B9S6Z
9ft1aznd8/fvaFwkVXXgZuzXCyJWdG7WiVwxk4JeUJYO7CShEFKf1W4b0HLVSFHf
Ns9bpEnFCGoIm+nEWHxfJIiKcKw48vOLBE7DLZZTyJFzuowSjA9uHXWPTSc3YW8Q
vmUEzvRA54+uezBZWK0YcxE+LqAfTaWzFjHbRaztxvf27XkQL+tNbBq8c3FCTEF2
ZQa0pP7IUM+wy3Lw7r8guZxzzhI1zU2MazFsb0R2z/mGPMdGDVF+kFZk6UYFm8Pm
aPM590xia5BCDOWAMoz/qzalR4KXrndxhfwSxtceMojAbok5VxOvuOdgOg==
-----END CERTIFICATE-----`))

	// Create tls.Config with desired tls properties
	return &tls.Config{
		RootCAs:    certpool,
		ServerName: "myquant",
		//ClientAuth: tls.NoClientCert,
	}
}

func TestTlsClient(t *testing.T) {
	var f MQTT.MessageHandler = func(c MQTT.Client, m MQTT.Message) {
		fmt.Println("rece:", m.Topic())
	}
	var ch chan int
	connOpts := &MQTT.ClientOptions{
		ClientID:             "ds-live",
		CleanSession:         true,
		Username:             "",
		Password:             "",
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            30 * time.Second,
		AutoReconnect:        true,
		PingTimeout:          10 * time.Second,
		ConnectTimeout:       30 * time.Second,
		//TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
		OnConnectionLost: func(c MQTT.Client, err error) { fmt.Println("mqtt disconnected.", zap.Error(err)) },
	}
	connOpts.AddBroker("ssl://127.0.0.1:8080").SetTLSConfig(NewTLSConfig(
		`/home/wendong/mygopath/src/grpcDemo/keys/server.pem`,
		`/home/wendong/mygopath/src/grpcDemo/keys/server.pem`,
		`/home/wendong/mygopath/src/grpcDemo/keys/server.key`,
	))

	mc := MQTT.NewClient(connOpts)
	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("mqtt connection failed.", zap.Error(token.Error()))
		return
	}
	defer mc.Disconnect(12)

	if token := mc.Subscribe("ds", 0, f); token.Wait() && token.Error() != nil {
		fmt.Println("sub failed.", zap.Error(token.Error()))

	} else {
		t := token.(*MQTT.SubscribeToken)
		fmt.Println(t.Result())
	}

	<-ch
}
