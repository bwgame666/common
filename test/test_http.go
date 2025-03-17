package main

import (
	"fmt"
	"github.com/bwgame666/common/service"
	"github.com/valyala/fasthttp"
)

type HelloReq struct {
	Name   string `json:"name" bind:"required,min=3,max=20" msg:"name length should be 3 to 20"`
	Status int    `json:"status" bind:"required"`
	Desc   string `json:"desc" bind:"default=test"`
}

type HelloResp struct {
	Msg string `json:"msg" validate:"required"`
}

func HelloControl(ctx *fasthttp.RequestCtx, req *HelloReq) (code int, msg string, resp HelloResp) {
	fmt.Println(string(ctx.Path()))
	fmt.Println(string(ctx.Request.Header.Peek("token")))
	resp.Msg = fmt.Sprintf("hello world! %s, status: %d, desc: %s", req.Name, req.Status, req.Desc)
	return 200, "success", resp
}

func main() {

	middleWares := []service.MiddlewareFunc{
		//service.DecryptMiddleware,
	}

	httpServer := service.New(middleWares, "sdfasdfqca", false)
	httpServer.Get("/hello", HelloControl)
	httpServer.StartServer(":8081")
}
