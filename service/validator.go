package service

import (
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
	"reflect"
)

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
		err := libs.JsonUnmarshal([]byte(jsonStr), paramValue)
		if err != nil {
			fmt.Println("Error:", err)
			svr.response(ctx, data)
			return
		}

		// 2、请求参数校验
		req := reflect.ValueOf(reqPara).Elem().Interface()
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
