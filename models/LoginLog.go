package models

import (
	"standard-library/consts"
	"standard-library/utility"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

type LoginLog struct {
	CommStruct
	UserId  int64  `orm:"description(用户ID)"`
	LoginIp string `orm:"description(登录IP)"`
}

func (lg *LoginLog) SetUpdateTime() {
	lg.UpdateTime = uint64(time.Now().Unix())
}

func (lg *LoginLog) SetCreateTime() {
	lg.CreateTime = uint64(time.Now().Unix())
}

func init() {
	orm.RegisterModel(new(LoginLog))
}

func (lg *LoginLog) TableName() string {
	return "api_login_log"
}

func (lg *LoginLog) AddLog(ip string, userId int64) (errCode int, err error) {
	lg.UserId = userId
	lg.LoginIp = ip
	lg.SetCreateTime()

	db := utility.NewDB()
	_, err = db.Insert(lg)
	if err != nil {
		errCode = consts.DB_INSERT_FAILED
		logs.Error("[LoginLog][AddLog] Insert AddLog error", err)
		return
	}

	return
}
