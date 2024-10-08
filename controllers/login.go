package controllers

import (
	"api-login-proto/common"
	"api-login-proto/login"
	"context"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

type LoginController struct {
	login.UnimplementedActivityChangeLogServiceServer
	BaseController
}

// Login
//
//	@Title			活动详情
//	@Description	活动详情
//	@Success		200			{object}	web.M
//	@Param			ActivityId	query		int64	false	活动ID
//	@router			/login [post]
func (ctl *LoginController) Login(ctx context.Context, request *common.Request) (*common.Response, error) {
	logs.Error("adslkjaslkjsd")
	return ctl.Success(web.M{
		"Items": "list",
	})
}
