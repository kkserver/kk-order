package order

import (
	"database/sql"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
)

const OrderStatusNone = 0      // 未支付
const OrderStatusPay = 200     //已支付
const OrderStatusCancel = 300  //已取消
const OrderStatusTimeout = 400 //已超时
const OrderStatusRefund = 500  //已退款

/**
 * 订单
 */
type Order struct {
	Id          int64  `json:"id"`
	Uid         int64  `json:"uid"`         //用户ID
	Title       string `json:"title"`       //说明
	Status      int    `json:"status"`      //状态
	Type        string `json:"type"`        //类型
	Options     string `json:"options"`     //选项
	PayTime     int64  `json:"payTime"`     //支付时间
	Expires     int64  `json:"expires"`     //失效时间
	Value       int64  `json:"value"`       //订单金额
	PayValue    int64  `json:"payValue"`    //支付金额
	RefundValue int64  `json:"refundValue"` //退款金额

	PayType    string `json:"payType"`    //支付类型
	PayTradeNo string `json:"payTradeNo"` //支付订单号

	Ctime int64 `json:"ctime"`
}

type IOrderApp interface {
	app.IApp
	GetDB() (*sql.DB, error)
	GetPrefix() string
	GetOrderTable() *kk.DBTable
}
