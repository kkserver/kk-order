package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderTask struct {
	app.Task
	Id     int64       `json:"id"`
	Uid    interface{} `json:"uid"` //用户ID
	Result OrderTaskResult
}

func (task *OrderTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderTask) GetInhertType() string {
	return "order"
}

func (task *OrderTask) GetClientName() string {
	return "Order.Get"
}
