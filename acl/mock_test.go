package acl

import (
	"testing"
	"fmt"
)
var f = func(userName string) *AuthInfo {
	return &AuthInfo{
		1, nil,
	}
}

func TestRegister(t *testing.T) {
	m,err:=NewTopicAclManger(TopicNumAuthType, "ss", f)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m.CheckSub("sss"))
	fmt.Println(m.CheckSub("sss"))
	fmt.Println(m.CheckSub("ssxs"))

	m1,err:=NewTopicAclManger(TopicNumAuthType, "xs", f)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println(m1.CheckSub("sa"))
	fmt.Println(m1.CheckSub("ab"))
}
