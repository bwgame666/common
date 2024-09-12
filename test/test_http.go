package main

import (
	"fmt"
	"github.com/bwgame666/common/service"
	"github.com/valyala/fasthttp"
)

type HelloReq struct {
	Name string `json:"name" validate:"required,min=3,max=20"`
	Desc string `json:"desc" validate:""`
}

type HelloResp struct {
	Msg string `json:"msg" validate:"required"`
}

func HelloControl(ctx *fasthttp.RequestCtx, req *HelloReq) (resp HelloResp) {
	fmt.Println(string(ctx.Path()))
	fmt.Println(string(ctx.Request.Header.Peek("token")))
	fmt.Println(req.Name, req.Desc)
	resp.Msg = "hello world!"
	return resp
}

func main() {

	middleWares := []service.MiddlewareFunc{
		//service.DecryptMiddleware,
	}

	httpServer := service.New(middleWares, "sdfasdfqca", false)
	httpServer.Get("/hello", HelloControl)
	httpServer.StartServer(":8081")
}
