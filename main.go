package main

import (
	"api-login-proto/login"
	"service-login/controllers"
	"standard-library/grpc"
	initilize "standard-library/initialize"
	"standard-library/validation"
)

func main() {
	initilize.InitLogs()
	initilize.InitNacosConfig()
	initilize.InitRedis()
	initilize.InitDB()
	initilize.InitMail()
	validation.Init()

	srv := grpc.NewServer()
	login.RegisterUserLoginServiceServer(srv.Srv, &controllers.LoginController{})
	initilize.RunGRPC(srv)
	//Service = {"cloud-micro-grpc-login": "localhost:55001","cloud-micro-grpc-netcash": "localhost:55006","cloud-micro-grpc-package_channel": "localhost:55002","cloud-micro-grpc-data-analysis": "localhost:55003","cloud-micro-grpc-game": "localhost:55011","cloud-micro-grpc-player": "localhost:55010","cloud-micro-grpc-customer": "localhost:55004","cloud-micro-grpc-download": "localhost:55019","cloud-micro-grpc-app-monitor": "localhost:55018","cloud-micro-grpc-agent": "localhost:55009", "cloud-micro-grpc-gameapi": "localhost:55020"}
}
