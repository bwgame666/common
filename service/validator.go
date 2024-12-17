package service

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
	"io"
	"reflect"
)

func getRequestArgs(ctx *fasthttp.RequestCtx, paramValue interface{}) error {
	if ctx.IsGet() || ctx.IsDelete() {
		requestArg := ctx.QueryArgs()
		//fmt.Println("request args: ", requestArg)
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
		if string(ctx.Request.Header.Peek("Accept-Encoding")) == "gzip" {
			// 创建一个字节读取器
			gzReader := bytes.NewReader(requestBody)
			gz, err := gzip.NewReader(gzReader)
			if err != nil {
				fmt.Println(ctx, "Failed to create gzip reader: ", err)
				return err
			}
			defer gz.Close()

			// 读取解压后的数据
			requestBody, err = io.ReadAll(gz)
			if err != nil {
				fmt.Println(ctx, "Failed to read gzip body: ", err)
				return err
			}
		}
		//fmt.Println("Request body: ", requestBody)
		err := libs.JsonUnmarshal(requestBody, paramValue)
		if err != nil {
			fmt.Println("[getRequestArgs] Request body: ", requestBody)
			fmt.Println("[getRequestArgs] JsonUnmarshal Error:", err)
			return err
		}
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
		funcType := reflect.TypeOf(handle)
		funcValue := reflect.ValueOf(handle)
		paramType := funcType.In(1)
		paramValue := reflect.New(paramType).Interface()

		// 1、读取请求参数
		err := getRequestArgs(ctx, &paramValue)
		if err != nil {
			fmt.Println("[validatorDecorator] getRequestArgs Error:", err)
			svr.Response(ctx, data)
			return
		}

		// 2、请求参数校验
		req := reflect.ValueOf(paramValue).Elem().Interface()
		validate := validator.New()
		errV := validate.Struct(req)
		if errV != nil {
			fmt.Println("[validatorDecorator] Validation errors:", errV)
			svr.Response(ctx, data)
			return
		}

		// 3、调用处理函数
		argValues := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
		returnValues := funcValue.Call(argValues)
		code := returnValues[0].Interface().(int)
		msg := returnValues[1].Interface().(string)
		d := returnValues[2].Interface()

		// 4、返回结果
		data.Code = code
		data.Message = msg
		data.Data = d
		svr.Response(ctx, data)
	}
}
