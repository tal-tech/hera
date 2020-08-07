package kafkaconsumer

import (
	"bytes"
	"fmt"
	logger "github.com/tal-tech/loggerX"
)

type Handler struct {

}
type HandlerFunc struct {
	MessageKey string
	Func       func([]byte) error
}

func(this*Handler)Deal(broker IBroker, partition int32, Offset int64, Key string, Value []byte, ext interface{}) bool {
	tag := "kafkaconsumer.Handler.Deal"
	pos := bytes.Index(Value, []byte(" "))
	pos_tab := bytes.Index(Value, []byte("\t"))
	if (pos <= 0 || pos_tab < pos) && pos_tab > 0 {
		pos = pos_tab
	}

	if pos > 0 {
		cmd := string(Value[0:pos])
		value := Value[pos+1:]
		if _, ok := execfuncs[cmd]; !ok {
			logger.D(tag, "不存在的注入方法,messageKey:%v", cmd)
			return true
		}
		err := execfuncs[cmd](value)
		if err != nil {
			return false
		}

	} else {
		info := fmt.Sprintf("p:%d,o:%d,k:%s,v:%s", partition, Offset, Key, string(Value))
		logger.E("INVALID_FORMAT", info)
		return true
	}

	return true
}
