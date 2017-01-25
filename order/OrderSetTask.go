package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderSetTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderSetTask struct {
	app.Task
	Id          int64       `json:"id"`
	Title       interface{} `json:"title"`       //说明
	Type        interface{} `json:"type"`        //类型
	Options     interface{} `json:"options"`     //选项
	Expires     interface{} `json:"expires"`     //失效时间
	Value       interface{} `json:"value"`       //订单金额
	PayValue    interface{} `json:"payValue"`    //支付金额
	RefundValue interface{} `json:"refundValue"` //退款金额
	Result      OrderSetTaskResult
}

func (task *OrderSetTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderSetTask) GetInhertType() string {
	return "order"
}

func (task *OrderSetTask) GetClientName() string {
	return "Order.Set"
}
