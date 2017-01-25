package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type TriggerOrderCancelDidTask struct {
	app.Task
	Order *Order `json:"order,omitempty"`
}

func (task *TriggerOrderCancelDidTask) GetResult() interface{} {
	return nil
}
