package acl

import (
	"fmt"
	"testing"
)

var f = func(userName string) *AuthInfo {
	return &AuthInfo{
		1,
		map[string]bool{"sa": true},
	}
}

func TestNewTopicAclManger(t *testing.T) {
	m, err := NewTopicAclManger(TopicNumAuthType, f("ss"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m.CheckSub("sss"))
	fmt.Println(m.CheckSub("sss"))
	fmt.Println(m.CheckSub("ssxs"))

	m1, err := NewTopicAclManger(TopicSetAuthType, f("kk"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m1.CheckSub("sa"))
	fmt.Println(m1.CheckSub("ab"))

	m2, err := NewTopicAclManger(TopicAlwaysVerifyType, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m2.CheckSub("sa"))
	fmt.Println(m2.CheckSub("ab"))

}
