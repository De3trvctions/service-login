package controllers

import (
	"api-login-proto/common"
	"api-login-proto/login"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"service-login/models"
	"standard-library/consts"
	"standard-library/jwt"
	"standard-library/models/dto"
	"standard-library/nacos"
	"standard-library/nets"
	"standard-library/redis"
	"standard-library/utility"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

type LoginController struct {
	login.UnimplementedUserLoginServiceServer
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
	req := dto.ReqLogin{}
	err := ctl.ParseJson(request, req)
	if err != nil {
		logs.Error("[LoginController][Login] Parse Json error", err)
		return ctl.Error(consts.PARAM_ERROR)
	}
	ableLogin, ableLoginRemaindingTime := ctl.getRedisLoginStatus(req.Username)

	if !ableLogin {
		leftSec := ableLoginRemaindingTime % 60
		leftMin := int(ableLoginRemaindingTime / 60)
		logs.Error("[LoginController][Login] 账号封锁中。剩余解封时间%d分钟%d秒", leftMin, leftSec)
		ctl.Error(consts.LOGIN_LOCK, fmt.Sprintf("请在%d分钟%d秒后再进行尝试", leftMin, leftSec))
	}

	acc := models.Account{}
	acc.Username = req.Username

	db := utility.NewDB()
	err = db.Get(&acc, "Username")
	if err != nil {
		logs.Error("[LoginController][Login] Account not found", err)
		ctl.Error(consts.USERNAME_NOT_FOUND)
	}

	hash := md5.Sum([]byte(req.Password))
	hashPassword := hex.EncodeToString(hash[:])
	if hashPassword != acc.Password {
		ctl.setRedisLoginFail(req.Username)
		logs.Error("[LoginController][Login] Password not match. req: %s", hashPassword)
		ctl.Error(consts.PASSWORD_NOT_MATCH)
	}

	// Generate JWT Token and return
	req.IP = nets.IP(ctl.Ctx.Request).String()
	token := getToken(req, acc.Id)

	ctl.delRedisLoginFail(req.Username)

	loginLog := models.LoginLog{}
	errCode, err := loginLog.AddLog(req.IP, acc.Id)
	if errCode != 0 || err != nil {
		logs.Error("[LoginController][Login] Add login log fail", err)
	}
	return ctl.Success(web.M{
		"Token":    token,
		"Username": acc.Username,
	})
}

func (ctl *LoginController) getRedisLoginStatus(username string) (ableLogin bool, remaindingTime int) {
	ableLogin = true
	ex1, _ := redis.Exists(fmt.Sprintf(consts.FailLoginAccountLock, username))

	if ex1 {
		ableLogin = false
		timeCache, _ := redis.Get(fmt.Sprintf(consts.FailLoginAccountLockTime, username))
		redisTime, _ := strconv.ParseInt(timeCache, 10, 64)
		timeLeft := time.Unix(redisTime, 0)
		remaindingTime = int(time.Until(timeLeft).Seconds())
	}
	return
}

func (ctl *LoginController) setRedisLoginFail(username string) (err error) {
	failCount := 0
	ex, _ := redis.Exists(fmt.Sprintf(consts.FailLoginCount, username))
	if ex {
		count, _ := redis.Get(fmt.Sprintf(consts.FailLoginCount, username))
		failCount, _ = strconv.Atoi(count)
		_, _ = redis.Del(fmt.Sprintf(consts.FailLoginCount, username))
	}

	if failCount >= 5 {
		_ = redis.Set(fmt.Sprintf(consts.FailLoginCount, username), failCount, time.Duration(15)*time.Minute)
		_ = redis.Set(fmt.Sprintf(consts.FailLoginAccountLock, username), 1, time.Duration(15)*time.Minute)
		_ = redis.Set(fmt.Sprintf(consts.FailLoginAccountLockTime, username), time.Now().Add(time.Duration(15)*time.Minute).Unix(), time.Duration(15)*time.Minute)
	} else {
		_ = redis.Set(fmt.Sprintf(consts.FailLoginCount, username), failCount)
	}
	return
}

func (ctl *LoginController) delRedisLoginFail(username string) {
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginCount, username))
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginAccountLock, username))
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginAccountLockTime, username))
}

func getToken(req dto.ReqLogin, id int64) (token string) {
	// tokenSalt, _ := web.AppConfig.String("TokenSalt")
	// tokenExpMinute, _ := web.AppConfig.Int("TokenExpMinute")
	delToken(req.Username)
	token = jwt.Gen(web.M{
		"Username":  req.Username,
		"AccountId": id,
		"Ip":        req.IP,
	}, nacos.TokenSalt, time.Duration(nacos.TokenExpMinute)*time.Minute)
	setToken(token, req.Username)
	return
}

func setToken(token, username string) {
	// Set token expired as 1 days
	_ = redis.Set(fmt.Sprintf(consts.AccountLoginByToken, token), username, time.Duration(nacos.TokenExpMinute)*time.Minute)
	_ = redis.Set(fmt.Sprintf(consts.AccountLoginByUsername, username), token, time.Duration(nacos.TokenExpMinute)*time.Minute)
}

func delToken(username string) {
	ex, _ := redis.Exists(fmt.Sprintf(consts.AccountLoginByUsername, username))
	if ex {
		token, _ := redis.Get(fmt.Sprintf(consts.AccountLoginByUsername, username))
		_, err1 := redis.Del(fmt.Sprintf(consts.AccountLoginByToken, token))
		_, err2 := redis.Del(fmt.Sprintf(consts.AccountLoginByUsername, username))
		if err1 != nil {
			logs.Error("[delToken] Error 1: ", err1)
		}
		if err2 != nil {
			logs.Error("[delToken] Error 2: ", err2)
		}
	}
}
