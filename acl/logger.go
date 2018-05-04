package acl

import (
	"fmt"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}
func SetLogger(l *zap.Logger) {
	if l == nil {
		fmt.Println(" acl.logger is nil !")
		l, _ = zap.NewDevelopment()
	}
	logger = l
}
