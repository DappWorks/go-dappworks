package vm

import "C"
import (
	"unsafe"

	"github.com/dappley/go-dappley/core/scState"
	logger "github.com/sirupsen/logrus"
)

//export TriggerEventFunc
func TriggerEventFunc(address unsafe.Pointer, topic *C.char, data *C.char) int {
	addr := uint64(uintptr(address))
	engine := getV8EngineByAddress(addr)
	t := C.GoString(topic)
	d := C.GoString(data)

	if engine == nil {
		logger.WithFields(logger.Fields{
			"contract_address": addr,
			"topic":            t,
			"data":             d,
		}).Debug("SmartContract: failed to get state handler!")
		return 1
	}

	engine.state.RecordEvent(scState.NewEvent(t, d))
	return 0
}
