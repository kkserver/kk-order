package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type TriggerOrderRefundDidTask struct {
	app.Task
	Order *Order `json:"order,omitempty"`
}

func (task *TriggerOrderRefundDidTask) GetResult() interface{} {
	return nil
}
