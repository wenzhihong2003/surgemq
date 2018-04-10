package acl

import "sync"

const (
	TopicNumAuthType = "topicNumAuth" //add userName
	TopicSetAuthType = "topicSetAuth"
)

type TopicNumAuth struct {
	Total       int
	NowTotal    int
	NowTotalMux sync.RWMutex
	TopicM      *sync.Map
}

var _ Authenticator = (*TopicNumAuth)(nil)

func NewTopicNumAuthProvider(total int) *TopicNumAuth {

	return &TopicNumAuth{
		TopicM: &sync.Map{},
		Total:  total,
	}
}

func (this *TopicNumAuth) CheckPub(topic string) bool {
	return true
}

func (this *TopicNumAuth) CheckSub(topic string) bool {
	this.NowTotalMux.Lock()
	defer this.NowTotalMux.Unlock()

	//total full
	if this.NowTotal > this.Total {
		return false
	}

	//not exists in map
	if _, ok := this.TopicM.Load(topic); !ok {
		if this.NowTotal == this.Total{
			return false
		}
		this.TopicM.Store(topic, true)
		this.NowTotal++
	}
	return true
}

func (this *TopicNumAuth) ProcessUnSub(topic string) {
	if _, ok := this.TopicM.Load(topic); !ok {
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

func NewTopicSetAuthProvider(topicMap map[string]bool) *TopicSetAuth {
	topicM := &sync.Map{}
	for k, v := range topicMap {
		topicM.Store(k, v)
	}
	return &TopicSetAuth{topicM}
}

func (this *TopicSetAuth) CheckPub(topic string) bool {
	return true
}

func (this *TopicSetAuth) CheckSub(topic string) bool {
	if _, ok := this.TopicM.Load(topic); !ok {
		return false
	}
	return true
}

func (this *TopicSetAuth) ProcessUnSub(topic string) {
	return
}
