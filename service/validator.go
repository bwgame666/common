package service

import (
	"fmt"
	"github.com/bwgame666/common/argsware"
	"github.com/valyala/fasthttp"
	"reflect"
)

func getRequestArgs(ctx *fasthttp.RequestCtx, paramValue interface{}) error {
	if ctx.IsGet() || ctx.IsDelete() {
		return argsware.QueryBindArgs(ctx, paramValue)
	} else if ctx.IsPost() || ctx.IsPut() {
		return argsware.JsonBindArgs(ctx, paramValue)
	}
	return nil
}

func validatorDecorator(svr *HttpService, handle RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		data := &ResponseData{
			Code:    408000,
			Message: "Parameter Invalid",
			Data:    make(map[string]interface{}),
		}

		// 获取 handle 的类型
		funcType := reflect.TypeOf(handle)
		if funcType.NumIn() < 2 {
			fmt.Println("[validatorDecorator] Error: handle must have at least two parameters")
			svr.Response(ctx, data)
			return
		}

		// 获取 handle 的第二个参数的类型
		paramType := funcType.In(1)
		if paramType.Kind() != reflect.Ptr || paramType.Elem().Kind() != reflect.Struct {
			fmt.Printf("[validatorDecorator] Error: the second parameter of handle must be a struct pointer, got: %v\n", paramType)
			svr.Response(ctx, data)
			return
		}

		funcValue := reflect.ValueOf(handle)

		// 创建该类型的实例
		var paramValue interface{}
		if paramType.Kind() == reflect.Ptr {
			// 如果 paramType 是指针类型，直接创建指向结构体的指针
			paramValue = reflect.New(paramType.Elem()).Interface()
		} else {
			// 如果 paramType 不是指针类型，创建指向结构体的指针
			paramValue = reflect.New(paramType).Interface()
		}

		// 1、读取请求参数并校验
		err := getRequestArgs(ctx, paramValue)
		if err != nil {
			fmt.Println("[validatorDecorator] getRequestArgs Error:", err)
			data.Message = err.Error()
			svr.Response(ctx, data)
			return
		}

		// 2、调用处理函数
		argValues := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(paramValue)}
		returnValues := funcValue.Call(argValues)
		code := returnValues[0].Interface().(int)
		msg := returnValues[1].Interface().(string)
		d := returnValues[2].Interface()

		// 3、返回结果
		data.Code = code
		data.Message = msg
		data.Data = d
		svr.Response(ctx, data)
	}
}
