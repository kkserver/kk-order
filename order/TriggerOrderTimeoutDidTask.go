package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type TriggerOrderTimeoutDidTask struct {
	app.Task
	Order *Order `json:"order,omitempty"`
}

func (task *TriggerOrderTimeoutDidTask) GetResult() interface{} {
	return nil
}
