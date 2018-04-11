package acl

import (
	"fmt"
	"testing"
)

var f = func(userName string) interface{} {
	return 1
}

var g = func(userName string) interface{} {
	return map[string]bool{
		"sa": true,
	}
}

func TestNewTopicAclManger(t *testing.T) {
	m, err := NewTopicAclManger(TopicNumAuthType, f("ss"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m.CheckSub([]byte("sss")))
	fmt.Println(m.CheckSub([]byte("sss")))
	fmt.Println(m.CheckSub([]byte("ssxs")))

	m1, err := NewTopicAclManger(TopicSetAuthType, g("kk"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m1.CheckSub([]byte("sa")))
	fmt.Println(m1.CheckSub([]byte("ab")))

	m2, err := NewTopicAclManger(TopicAlwaysVerifyType, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m2.CheckSub([]byte("sa")))
	fmt.Println(m2.CheckSub([]byte("ab")))

	m3, err := NewTopicAclManger(TopicSetAuthType, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m3.CheckSub([]byte("sa")))
	fmt.Println(m3.CheckSub([]byte("ab")))

	m4, err := NewTopicAclManger(TopicNumAuthType, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m4.CheckSub([]byte("sss")))
	fmt.Println(m4.CheckSub([]byte("sss")))
	fmt.Println(m4.CheckSub([]byte("ssxs")))

}
