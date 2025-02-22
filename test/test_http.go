package main

import (
	"fmt"
	"github.com/bwgame666/common/service"
	"github.com/valyala/fasthttp"
)

type HelloReq struct {
	Name   string `json:"name" rules:"required,rule=str,min=3,max=20" msg:"name length should be 3 to 20"`
	Status int    `json:"status" rules:"required,rule=str,min=0,max=1" msg:"错误的state, [0,1]"`
	Desc   string `json:"desc"`
}

type HelloResp struct {
	Msg string `json:"msg" validate:"required"`
}

func HelloControl(ctx *fasthttp.RequestCtx, req *HelloReq) (code int, msg string, resp HelloResp) {
	err := service.ValidateArgs(req)
	if err != nil {
		resp.Msg = err.Error()
		return 408000, "invalid", resp
	}

	fmt.Println(string(ctx.Path()))
	fmt.Println(string(ctx.Request.Header.Peek("token")))
	resp.Msg = "hello world!"
	return 200, "success", resp
}

func main() {

	middleWares := []service.MiddlewareFunc{
		//service.DecryptMiddleware,
	}

	httpServer := service.New(middleWares, "sdfasdfqca", false)
	httpServer.Post("/hello", HelloControl)
	httpServer.StartServer(":8081")
}
