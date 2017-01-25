package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderRefundTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderRefundTask struct {
	app.Task
	Id     int64       `json:"id"`
	Value  interface{} `json:"value"` //退款金额
	Result OrderRefundTaskResult
}

func (task *OrderRefundTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderRefundTask) GetInhertType() string {
	return "order"
}

func (task *OrderRefundTask) GetClientName() string {
	return "Order.Refund"
}
