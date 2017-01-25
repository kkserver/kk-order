package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type TriggerOrderPayDidTask struct {
	app.Task
	Order *Order `json:"order,omitempty"`
}

func (task *TriggerOrderPayDidTask) GetResult() interface{} {
	return nil
}
