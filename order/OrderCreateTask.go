package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderCreateTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderCreateTask struct {
	app.Task
	Uid         int64       `json:"uid"`         //用户ID
	Title       string      `json:"title"`       //说明
	Type        string      `json:"type"`        //类型
	Options     interface{} `json:"options"`     //选项
	Expires     int64       `json:"expires"`     //失效时间
	Value       int64       `json:"value"`       //订单金额
	PayValue    int64       `json:"payValue"`    //支付金额
	RefundValue int64       `json:"refundValue"` //退款金额
	Result      OrderCreateTaskResult
}

func (task *OrderCreateTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderCreateTask) GetInhertType() string {
	return "order"
}

func (task *OrderCreateTask) GetClientName() string {
	return "Order.Create"
}
