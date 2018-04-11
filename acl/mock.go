package acl

import "sync"

const (
	TopicNumAuthType      = "topicNumAuth"
	TopicSetAuthType      = "topicSetAuth"
	TopicAlwaysVerifyType = "topicAlwaysVerify"
)

type TopicNumAuth struct {
	Total       int
	NowTotal    int
	NowTotalMux sync.RWMutex
	TopicM      sync.Map
}

var _ Authenticator = (*TopicNumAuth)(nil)

func NewTopicNumAuthProvider(i interface{}) *TopicNumAuth {
	var total int
	if v, ok := i.(int); ok {
		total = v
	}

	return &TopicNumAuth{
		TopicM: sync.Map{},
		Total:  total,
	}
}

func (this *TopicNumAuth) CheckPub(topic []byte) bool {
	return true
}

func (this *TopicNumAuth) CheckSub(topic []byte) bool {
	this.NowTotalMux.Lock()
	defer this.NowTotalMux.Unlock()

	//total full
	if this.NowTotal > this.Total {
		return false
	}

	//not exists in map
	if _, ok := this.TopicM.Load(string(topic)); !ok {
		if this.NowTotal == this.Total {
			return false
		}
		this.TopicM.Store(string(topic), true)
		this.NowTotal++
	}
	return true
}

func (this *TopicNumAuth) ProcessUnSub(topic []byte) {
	if _, ok := this.TopicM.Load(string(topic)); !ok {
		return
	}
	this.NowTotalMux.Lock()
	defer this.NowTotalMux.Unlock()

	this.NowTotal--
}

type TopicSetAuth struct {
	TopicM *sync.Map
}

var _ Authenticator = (*TopicSetAuth)(nil)

func NewTopicSetAuthProvider(i interface{}) *TopicSetAuth {
	topicM := &sync.Map{}
	topicMap := make(map[string]bool)
	if v, ok := i.(map[string]bool); ok {
		topicMap = v
	}
	for k, v := range topicMap {
		topicM.Store(k, v)
	}
	return &TopicSetAuth{topicM}
}

func (this *TopicSetAuth) CheckPub(topic []byte) bool {
	return true
}

func (this *TopicSetAuth) CheckSub(topic []byte) bool {
	if _, ok := this.TopicM.Load(string(topic)); !ok {
		return false
	}
	return true
}

func (this *TopicSetAuth) ProcessUnSub(topic []byte) {
	return
}

type TopicAlwaysVerify bool

var _ Authenticator = (*TopicAlwaysVerify)(nil)

func (this TopicAlwaysVerify) CheckPub(topic []byte) bool {
	return true

}

func (this TopicAlwaysVerify) CheckSub(topic []byte) bool {
	return true

}

func (this TopicAlwaysVerify) ProcessUnSub(topic []byte) {
}
