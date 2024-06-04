package service

import (
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
	"reflect"
)

func getRequestArgs(ctx *fasthttp.RequestCtx, paramValue interface{}) error {
	if ctx.IsGet() || ctx.IsDelete() {
		requestArg := ctx.QueryArgs()
		fmt.Println("request args: ", requestArg)
		requestMap := make(map[string]interface{})
		requestArg.VisitAll(func(key, value []byte) {
			requestMap[string(key)] = string(value)
		})
		jsonB, _ := sonic.Marshal(requestMap)
		err := libs.JsonUnmarshal(jsonB, paramValue)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
	} else if ctx.IsPost() || ctx.IsPut() {
		// 读取post请求参数
		requestBody := ctx.PostBody()
		fmt.Println("Request body: ", requestBody)
		err := libs.JsonUnmarshal(requestBody, paramValue)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
	}
	return nil
}

func validatorDecorator(svr *HttpService, handle RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		data := &ResponseData{
			Code:    408002,
			Message: "Parameter Invalid",
		}
		funcType := reflect.TypeOf(handle)
		funcValue := reflect.ValueOf(handle)
		paramType := funcType.In(0)
		paramValue := reflect.New(paramType).Interface()

		// 1、读取请求参数
		err := getRequestArgs(ctx, &paramValue)
		if err != nil {
			fmt.Println("Error:", err)
			svr.response(ctx, data)
			return
		}

		// 2、请求参数校验
		req := reflect.ValueOf(paramValue).Elem().Interface()
		validate := validator.New()
		errV := validate.Struct(req)
		if errV != nil {
			fmt.Println("Validation errors:", errV)
			svr.response(ctx, data)
			return
		}

		// 3、调用处理函数
		argValues := []reflect.Value{reflect.ValueOf(req)}
		returnValues := funcValue.Call(argValues)
		resp := returnValues[0].Interface()

		// 4、返回结果
		data.Code = 200
		data.Message = "success"
		data.Data = resp
		svr.response(ctx, data)
	}
}
