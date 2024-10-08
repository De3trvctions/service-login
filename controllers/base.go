package controllers

import (
	"api-login-proto/common"
	"context"
	"errors"
	"fmt"
	"standard-library/consts"
	"standard-library/grpc"
	"standard-library/json"

	"github.com/beego/beego/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"
)

type BaseController struct {
	web.Controller
	i18n.Locale

	GrpcConn        grpc.Conn
	GrpcCtx         context.Context
	GrpcServiceName string
}

func init() {}

// CheckLanguage 检测语言包
func (ctl *BaseController) CheckLanguage() {
	ctl.Lang = "zh-CN"
}

// Prepare 在这里处理后，其他函数中就不需要雷同代码了。
func (ctl *BaseController) Prepare() {
	ctl.StartTimeCreate()
	logs.Debug("BaseController Prepare")
}

func (ctl *BaseController) SetRequest(request *common.Request) {
	request = request
}

// StartTimeCreate 路由匹配之前的时间
func (ctl *BaseController) StartTimeCreate() {

}

// GetSuccess 专用cloud this.Success()返回，替换为grpc模式返回
func (ctl *BaseController) GetSuccess(obj interface{}) string {
	return ctl.TraceJsonGrpc(consts.SUCCESS_REQUEST, "", obj)
}

// 返回错误
func (ctl *BaseController) Error(code int, msg ...string) (*common.Response, error) {
	data := ""
	if len(msg) > 0 {
		data = fmt.Sprint(msg)
	}
	return &common.Response{
		Code: int64(code),
		Data: data,
	}, nil

}

// Success 返回成功
func (ctl *BaseController) Success(obj interface{}) (*common.Response, error) {
	return &common.Response{
		Code: consts.SUCCESS_REQUEST,
		Data: ctl.GetSuccess(obj),
	}, nil
}

// GetError 专用cloud this.Error()返回，替换为grpc模式返回
func (ctl *BaseController) GetError(code int, msg ...string) string {
	if code < 1 {
		code = consts.SERVER_ERROR
	}

	resMsg := ""
	if web.BConfig.RunMode == web.DEV && len(msg) > 0 {
		resMsg = fmt.Sprint(msg)
	} else {
		ctl.CheckLanguage()
		if code > 0 {
			resMsg = ctl.Tr("error." + fmt.Sprintf("%d", code))
		}
	}

	return ctl.TraceJsonGrpc(int64(code), resMsg, nil)
}

// TraceJsonGrpc 单服务，直接兼容返回结果
func (ctl *BaseController) TraceJsonGrpc(code int64, msg string, result interface{}) string {
	data := map[string]interface{}{"Code": code, "Msg": msg, "Data": result}
	res, err := json.StringifyE(data)
	if err != nil {
		logs.Error("[DirectTraceJson]解析失败", err, ctl.Data["json"])
		res, _ = json.StringifyE(map[string]interface{}{"Code": "100000", "Msg": "解析失败", "Data": ""})
		return string(res)
	}

	return string(res)
}
func (ctl *BaseController) ParseJson(request *common.Request, r interface{}) (err error) {
	err = json.ParseE(request.Data, r)
	return
}

// 创建GRPC连接
func (ctl *BaseController) ConnGRpc(serviceName string) error {
	ctl.GrpcServiceName = serviceName
	var err error
	ctl.GrpcConn, err = grpc.Get(serviceName)
	if err != nil {
		logs.Error("[controllers>ConnGRpc]获取服务[%s]失败 %s", serviceName, err)
		return err
	}

	// 针对异常进行处理
	if ctl.GrpcConn == nil {
		logs.Error("[controllers>ConnGRpc]获取服务[%s]conn为空，重新获取服务", serviceName)
		if err == nil {
			return errors.New("conn is nil")
		}
		ctl.GrpcConn, _ = grpc.Get(serviceName)
	}

	return nil
}
