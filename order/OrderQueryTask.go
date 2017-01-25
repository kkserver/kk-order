package order

import (
	"github.com/kkserver/kk-lib/kk/app"
)

type OrderQueryCounter struct {
	PageIndex int `json:"p"`
	PageSize  int `json:"size"`
	PageCount int `json:"count"`
}

type OrderQueryTaskResult struct {
	app.Result
	Counter *OrderQueryCounter `json:"counter,omitempty"`
	Orders  []Order            `json:"orders,omitempty"`
}

type OrderQueryTask struct {
	app.Task
	Id        int64  `json:"id"`
	Uid       int64  `json:"uid"` //用户ID
	Prefix    string `json:"prefix"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	OrderBy   string `json:"orderBy"` // desc, asc
	PageIndex int    `json:"p"`
	PageSize  int    `json:"size"`
	Counter   bool   `json:"counter"`
	Result    OrderQueryTaskResult
}

func (task *OrderQueryTask) GetResult() interface{} {
	return &task.Result
}

func (task *OrderQueryTask) GetInhertType() string {
	return "order"
}

func (task *OrderQueryTask) GetClientName() string {
	return "Order.Query"
}
