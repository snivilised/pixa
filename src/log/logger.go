package log

import (
	"github.com/snivilised/pixa/src/internal/log"
)

type Logger interface {
	Debug(msg string, fields ...log.Field)
	Info(msg string, fields ...log.Field)
	Warn(msg string, fields ...log.Field)
	Error(msg string, fields ...log.Field)
}