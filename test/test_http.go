package main

import (
	"fmt"
	"github.com/bwgame666/common/argsware"
	"github.com/bwgame666/common/service"
	"github.com/valyala/fasthttp"
	"slices"
	"strconv"
	"strings"
)

type HelloReq struct {
	Name   string `json:"name" bind:"required,min=3,max=20" msg:"name length should be 3 to 20"`
	Status []int  `json:"status" bind:"required,rule=status,min=1,max=3" arg:"10,11,12,13,14,15"`
	DescID string `json:"desc,omitempty" bind:"required"`
}

type HelloResp struct {
	Msg string `json:"msg" validate:"required"`
}

func HelloControl(ctx *fasthttp.RequestCtx, req *HelloReq) (code int, msg string, resp HelloResp) {
	fmt.Println(string(ctx.Path()))
	fmt.Println(string(ctx.Request.Header.Peek("token")))
	resp.Msg = fmt.Sprintf("hello world! %s, status: %d, desc: %s", req.Name, req.Status, req.DescID)
	return 200, "success", resp
}

func CheckStatus(fieldValue interface{}, fieldName, paramArg string) error {
	value, ok := fieldValue.([]int)
	if !ok {
		return argsware.NewArgsError("CheckStatus", fieldName, "the fieldValue must be a []int")
	}

	params := strings.Split(paramArg, ",")
	for _, v := range value {
		if !slices.Contains(params, strconv.Itoa(v)) {
			return argsware.NewArgsError("CheckStatus", fieldName, fmt.Sprintf("%d is not in %s", v, paramArg))
		}
	}

	return nil
}

func main() {

	middleWares := []service.MiddlewareFunc{
		//service.DecryptMiddleware,
	}
	argsware.Register("status", CheckStatus)

	httpServer := service.New(middleWares, "sdfasdfqca", false)
	httpServer.Get("/hello", HelloControl)
	httpServer.StartServer(":8081")
}
