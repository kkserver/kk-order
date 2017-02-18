package order

import (
	"bytes"
	"fmt"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/dynamic"
	"github.com/kkserver/kk-lib/kk/json"
	"log"
	"strings"
	"time"
)

type OrderService struct {
	app.Service

	Create *OrderCreateTask
	Set    *OrderSetTask
	Get    *OrderTask
	Pay    *OrderPayTask
	Cancel *OrderCancelTask
	Refund *OrderRefundTask
	Query  *OrderQueryTask
}

func (S *OrderService) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func (S *OrderService) HandleRunloopTask(a IOrderApp, task *app.RunloopTask) error {

	var db, err = a.GetDB()

	if err != nil {
		return err
	}

	var fn func() = nil

	fn = func() {

		now := time.Now().Unix()

		for {

			count := 0

			rows, err := kk.DBQuery(db, a.GetOrderTable(), a.GetPrefix(), " WHERE status=? AND ctime + expires <= ?  ORDER BY id ASC LIMIT 1", OrderStatusNone, now)

			if err != nil {
				log.Println("OrderService", "Runloop", "Fail", err.Error())
			} else {

				v := Order{}
				scanner := kk.NewDBScaner(&v)

				if rows.Next() {

					err = scanner.Scan(rows)

					rows.Close()

					if err != nil {
						log.Println("OrderService", "Runloop", "Fail", err.Error())
					} else {

						count = count + 1

						v.Status = OrderStatusTimeout

						_, err = kk.DBUpdateWithKeys(db, a.GetOrderTable(), a.GetPrefix(), &v, map[string]bool{"status": true})

						if err != nil {
							log.Println("OrderService", "Runloop", "Fail", err.Error())
						} else {

							did := TriggerOrderTimeoutDidTask{}
							did.Order = &v

							err = app.Handle(a, &did)

							if err != nil {
								log.Println("OrderService", "Runloop", "Fail", err.Error())
							}

						}

					}

				} else {
					rows.Close()
				}

			}

			if count == 0 {
				break
			}
		}

		log.Println("OrderService", "Runloop", "OK")

		a.GetRunloop().AsyncDelay(fn, 10*time.Second)

	}

	fn()

	return nil
}

func (S *OrderService) HandleOrderCreateTask(a IOrderApp, task *OrderCreateTask) error {

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	v := Order{}

	v.Uid = task.Uid
	v.Title = task.Title
	v.Type = task.Type

	b, err := json.Encode(task.Options)

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	v.Options = string(b)
	v.Expires = task.Expires
	v.Value = task.Value
	v.PayValue = task.PayValue
	v.RefundValue = task.RefundValue
	v.Ctime = time.Now().Unix()

	_, err = kk.DBInsert(db, a.GetOrderTable(), a.GetPrefix(), &v)

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	{
		did := TriggerOrderCreateDidTask{}
		did.Order = &v

		err := app.Handle(a, &did)

		if err != nil {
			e, ok := err.(*app.Error)
			if ok {
				task.Result.Errno = e.Errno
				task.Result.Errmsg = e.Errmsg
			} else {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
			}
			return nil
		}
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderSetTask(a IOrderApp, task *OrderSetTask) error {

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	v := Order{}

	rows, err := kk.DBQuery(db, a.GetOrderTable(), a.GetPrefix(), " WHERE id=?", task.Id)

	if err != nil {
		return err
	}

	defer rows.Close()

	if rows.Next() {

		scanner := kk.NewDBScaner(&v)

		err = scanner.Scan(rows)

		if err != nil {
			if err != nil {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
				return nil
			}
		}

		keys := map[string]bool{}

		if task.Title != nil {
			v.Title = dynamic.StringValue(task.Title, v.Title)
			keys["title"] = true
		}

		if task.Type != nil {
			v.Type = dynamic.StringValue(task.Type, v.Type)
			keys["type"] = true
		}

		if task.Options != nil {
			b, err := json.Encode(task.Options)
			if err != nil {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
				return nil
			}
			v.Options = string(b)
			keys["options"] = true
		}

		if task.Expires != nil {
			v.Expires = dynamic.IntValue(task.Expires, v.Expires)
			keys["expires"] = true
		}

		if task.Value != nil {
			v.Value = dynamic.IntValue(task.Value, v.Value)
			keys["value"] = true
		}

		if task.PayValue != nil {
			v.PayValue = dynamic.IntValue(task.PayValue, v.PayValue)
			keys["payvalue"] = true
		}

		if task.RefundValue != nil {
			v.RefundValue = dynamic.IntValue(task.RefundValue, v.RefundValue)
			keys["refundvalue"] = true
		}

		_, err = kk.DBUpdateWithKeys(db, a.GetOrderTable(), a.GetPrefix(), &v, keys)

		if err != nil {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}

	} else {
		return app.NewError(ERROR_ORDER_NOT_FOUND, "Not Found Order")
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderCancelTask(a IOrderApp, task *OrderCancelTask) error {

	if task.Id == 0 {
		task.Result.Errno = ERROR_ORDER_NOT_FOUND_ID
		task.Result.Errmsg = "Not found id"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	tx, err := db.Begin()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var v = Order{}

	err = func() error {

		rows, err := kk.DBQuery(tx, a.GetOrderTable(), a.GetPrefix(), " WHERE id=? FOR UPDATE", task.Id)

		if err != nil {
			return err
		}

		if rows.Next() {

			scanner := kk.NewDBScaner(&v)

			err = scanner.Scan(rows)

			rows.Close()

			if err != nil {
				return err
			}

			if v.Status != OrderStatusNone {
				return app.NewError(ERROR_ORDER_STATUS, "The current state can not be modified")
			}

			v.Status = OrderStatusCancel

			_, err = kk.DBUpdateWithKeys(tx, a.GetOrderTable(), a.GetPrefix(), &v, map[string]bool{"status": true})

			if err != nil {
				return err
			}

		} else {
			rows.Close()
			return app.NewError(ERROR_ORDER_NOT_FOUND, "Not Found Order")
		}

		return nil
	}()

	if err == nil {
		err = tx.Commit()
	}

	if err != nil {
		tx.Rollback()
		e, ok := err.(*app.Error)
		if ok {
			task.Result.Errno = e.Errno
			task.Result.Errmsg = e.Errmsg
			return nil
		} else {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}
	}

	{
		did := TriggerOrderCancelDidTask{}
		did.Order = &v

		err := app.Handle(a, &did)

		if err != nil {
			e, ok := err.(*app.Error)
			if ok {
				task.Result.Errno = e.Errno
				task.Result.Errmsg = e.Errmsg
			} else {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
			}
			return nil
		}
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderPayTask(a IOrderApp, task *OrderPayTask) error {

	if task.Id == 0 {
		task.Result.Errno = ERROR_ORDER_NOT_FOUND_ID
		task.Result.Errmsg = "Not found id"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	tx, err := db.Begin()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var v = Order{}

	err = func() error {

		rows, err := kk.DBQuery(tx, a.GetOrderTable(), a.GetPrefix(), " WHERE id=? FOR UPDATE", task.Id)

		if err != nil {
			return err
		}

		if rows.Next() {

			scanner := kk.NewDBScaner(&v)

			err = scanner.Scan(rows)

			rows.Close()

			if err != nil {
				return err
			}

			if v.Status != OrderStatusNone {
				return app.NewError(ERROR_ORDER_STATUS, "The current state can not be modified")
			}

			v.Status = OrderStatusPay

			if task.Value != nil {
				v.PayValue = dynamic.IntValue(task.Value, v.PayValue)
			}

			v.PayType = task.PayType
			v.PayTradeNo = task.PayTradeNo

			_, err = kk.DBUpdateWithKeys(tx, a.GetOrderTable(), a.GetPrefix(), &v, map[string]bool{"status": true, "payvalue": true, "paytype": true, "paytradeno": true})

			if err != nil {
				return err
			}

		} else {
			rows.Close()
			return app.NewError(ERROR_ORDER_NOT_FOUND, "Not Found Order")
		}

		return nil
	}()

	if err == nil {
		err = tx.Commit()
	}

	if err != nil {
		tx.Rollback()
		e, ok := err.(*app.Error)
		if ok {
			task.Result.Errno = e.Errno
			task.Result.Errmsg = e.Errmsg
			return nil
		} else {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}
	}

	{
		did := TriggerOrderPayDidTask{}
		did.Order = &v

		err := app.Handle(a, &did)

		if err != nil {
			e, ok := err.(*app.Error)
			if ok {
				task.Result.Errno = e.Errno
				task.Result.Errmsg = e.Errmsg
			} else {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
			}
			return nil
		}
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderRefundTask(a IOrderApp, task *OrderRefundTask) error {

	if task.Id == 0 {
		task.Result.Errno = ERROR_ORDER_NOT_FOUND_ID
		task.Result.Errmsg = "Not found id"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	tx, err := db.Begin()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var v = Order{}

	err = func() error {

		rows, err := kk.DBQuery(tx, a.GetOrderTable(), a.GetPrefix(), " WHERE id=? FOR UPDATE", task.Id)

		if err != nil {
			return err
		}

		if rows.Next() {

			scanner := kk.NewDBScaner(&v)

			err = scanner.Scan(rows)

			rows.Close()

			if err != nil {
				return err
			}

			if v.Status != OrderStatusPay {
				return app.NewError(ERROR_ORDER_STATUS, "The current state can not be modified")
			}

			if task.Value != nil {
				v.RefundValue = dynamic.IntValue(task.Value, v.RefundValue)
			}

			v.Status = OrderStatusRefund

			_, err = kk.DBUpdateWithKeys(tx, a.GetOrderTable(), a.GetPrefix(), &v, map[string]bool{"status": true, "refundvalue": true})

			if err != nil {
				return err
			}

		} else {
			rows.Close()
			return app.NewError(ERROR_ORDER_NOT_FOUND, "Not Found Order")
		}

		return nil
	}()

	if err == nil {
		err = tx.Commit()
	}

	if err != nil {
		tx.Rollback()
		e, ok := err.(*app.Error)
		if ok {
			task.Result.Errno = e.Errno
			task.Result.Errmsg = e.Errmsg
			return nil
		} else {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}
	}

	{
		did := TriggerOrderPayDidTask{}
		did.Order = &v

		err := app.Handle(a, &did)

		if err != nil {
			e, ok := err.(*app.Error)
			if ok {
				task.Result.Errno = e.Errno
				task.Result.Errmsg = e.Errmsg
			} else {
				task.Result.Errno = ERROR_ORDER
				task.Result.Errmsg = err.Error()
			}
			return nil
		}
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderTask(a IOrderApp, task *OrderTask) error {

	if task.Id == 0 {
		task.Result.Errno = ERROR_ORDER_NOT_FOUND_ID
		task.Result.Errmsg = "Not found id"
		return nil
	}

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var args = []interface{}{}

	var sql = bytes.NewBuffer(nil)

	sql.WriteString(" WHERE id=?")

	args = append(args, task.Id)

	if task.Uid != nil {
		sql.WriteString(" AND uid=?")
		args = append(args, task.Uid)
	}

	rows, err := kk.DBQuery(db, a.GetOrderTable(), a.GetPrefix(), sql.String(), args...)

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	v := Order{}

	if rows.Next() {

		scanner := kk.NewDBScaner(&v)

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}

	} else {
		return app.NewError(ERROR_ORDER_NOT_FOUND, "Not Found Order")
	}

	task.Result.Order = &v

	return nil
}

func (S *OrderService) HandleOrderQueryTask(a IOrderApp, task *OrderQueryTask) error {

	var db, err = a.GetDB()

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	var orders = []Order{}

	var args = []interface{}{}

	var sql = bytes.NewBuffer(nil)

	sql.WriteString(" WHERE 1")

	if task.Id != 0 {
		sql.WriteString(" AND id=?")
		args = append(args, task.Id)
	}

	if task.Prefix != "" {
		sql.WriteString(" AND `type` LIKE ?")
		args = append(args, task.Prefix+"%")
	}

	if task.Type != "" {
		sql.WriteString(" AND `type` = ?")
		args = append(args, task.Type)
	}

	if task.Uid != 0 {
		sql.WriteString(" AND uid = ?")
		args = append(args, task.Uid)
	}

	if task.Status != "" {
		vs := strings.Split(task.Status, ",")
		sql.WriteString(" AND status IN (")
		for i, v := range vs {
			if i != 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			args = append(args, v)
		}
		sql.WriteString(")")
	}

	if task.OrderBy == "asc" {
		sql.WriteString(" ORDER BY id ASC")
	} else {
		sql.WriteString(" ORDER BY id DESC")
	}

	var pageIndex = task.PageIndex
	var pageSize = task.PageSize

	if pageIndex < 1 {
		pageIndex = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	if task.Counter {
		var counter = OrderQueryCounter{}
		counter.PageIndex = pageIndex
		counter.PageSize = pageSize
		counter.RowCount, err = kk.DBQueryCount(db, a.GetOrderTable(), a.GetPrefix(), sql.String(), args...)
		if err != nil {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}
		if counter.RowCount%pageSize == 0 {
			counter.PageCount = counter.RowCount / pageSize
		} else {
			counter.PageCount = counter.RowCount/pageSize + 1
		}
		task.Result.Counter = &counter
	}

	sql.WriteString(fmt.Sprintf(" LIMIT %d,%d", (pageIndex-1)*pageSize, pageSize))

	log.Println("SQL", sql.String())

	var v = Order{}
	var scanner = kk.NewDBScaner(&v)

	rows, err := kk.DBQuery(db, a.GetOrderTable(), a.GetPrefix(), sql.String(), args...)

	if err != nil {
		task.Result.Errno = ERROR_ORDER
		task.Result.Errmsg = err.Error()
		return nil
	}

	defer rows.Close()

	for rows.Next() {

		err = scanner.Scan(rows)

		if err != nil {
			task.Result.Errno = ERROR_ORDER
			task.Result.Errmsg = err.Error()
			return nil
		}

		orders = append(orders, v)
	}

	task.Result.Orders = orders

	return nil
}
