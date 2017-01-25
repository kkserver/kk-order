package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type TriggerOrderCreateDidTask struct {
	app.Task
	Order *Order `json:"order,omitempty"`
}

func (task *TriggerOrderCreateDidTask) GetResult() interface{} {
	return nil
}
