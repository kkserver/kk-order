package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderCancelTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderCancelTask struct {
	app.Task
	Id     int64 `json:"id"`
	Result OrderCancelTaskResult
}

func (task *OrderCancelTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderCancelTask) GetInhertType() string {
	return "order"
}

func (task *OrderCancelTask) GetClientName() string {
	return "Order.Cancel"
}
