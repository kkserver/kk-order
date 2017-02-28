package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderPayTaskResult struct {
	app.Result
	Order *Order `json:"order,omitempty"`
}

type OrderPayTask struct {
	app.Task
	Id         int64       `json:"id"`
	Value      interface{} `json:"value"`      //支付金额
	PayType    string      `json:"payType"`    //支付类型
	PayTradeNo string      `json:"payTradeNo"` //支付订单号
	Options    interface{} `json:"options"`    //选项
	Result     OrderPayTaskResult
}

func (task *OrderPayTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderPayTask) GetInhertType() string {
	return "order"
}

func (task *OrderPayTask) GetClientName() string {
	return "Order.Pay"
}
