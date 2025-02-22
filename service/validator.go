package service

import (
	"errors"
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/valyala/fasthttp"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func getRequestArgs(ctx *fasthttp.RequestCtx, paramValue interface{}) error {
	if ctx.IsGet() || ctx.IsDelete() {
		requestArg := ctx.QueryArgs()
		requestMap := make(map[string]interface{})
		requestArg.VisitAll(func(key, value []byte) {
			requestMap[string(key)] = string(value)
		})
		jsonB, _ := sonic.Marshal(requestMap)
		err := libs.JsonUnmarshal(jsonB, paramValue)
		if err != nil {
			fmt.Println("[getRequestArgs] request args: ", requestArg)
			fmt.Println("[getRequestArgs] Error:", err)
			return err
		}
	} else if ctx.IsPost() || ctx.IsPut() {
		// 读取post请求参数
		requestBody := ctx.PostBody()
		//fmt.Println("Request body: ", requestBody)
		err := libs.JsonUnmarshal(requestBody, paramValue)
		if err != nil {
			fmt.Println("[getRequestArgs] Request body: ", string(requestBody))
			fmt.Println("[getRequestArgs] JsonUnmarshal Error:", err)
			return err
		}
	}
	return nil
}

func ValidateArgs(objs interface{}) error {
	rt := reflect.ValueOf(objs)
	if rt.Kind() != reflect.Ptr {
		return errors.New("argument 2 should be map or ptr")
	}

	rtElem := rt.Elem()
	if rtElem.Kind() != reflect.Struct {
		return errors.New("non-structure type not supported yet")
	}

	s := rtElem
	fmt.Println("s", s.NumField(), s)
	for i := 0; i < s.NumField(); i++ {

		f := s.Type().Field(i)

		min := int64(0)
		max := int64(math.MaxInt64)

		name := f.Tag.Get("json")
		name = strings.Split(name, ",")[0]
		if len(name) == 0 {
			name = strings.ToLower(f.Name)
		}

		msg := f.Tag.Get("msg")
		if len(msg) == 0 {
			msg = "Parameter Invalid"
		}

		rules := validate(f.Tag.Get("rules"))

		rule := getValue(rules, "rule")
		required := getValue(rules, "required")
		def := getValue(rules, "default")

		minStr := getValue(rules, "min")
		if len(minStr) > 0 {
			if v, err := strconv.ParseInt(minStr, 10, 64); err == nil {
				min = v
			}
		}
		maxStr := getValue(rules, "max")
		if len(maxStr) > 0 {
			if v, err := strconv.ParseInt(maxStr, 10, 64); err == nil {
				max = v
			}
		}

		defaultVal := ""
		if s.Field(i).CanInterface() {
			defaultVal = fmt.Sprintf("%v", s.Field(i).Interface())
		}

		nums := len(def)
		check := true //默认需要校验
		fmt.Println("name", name, "required", required, "default", defaultVal, "min", min, "max", max, "rule", rule)
		if defaultVal == "" {
			if nums > 0 {
				defaultVal = def
			}

			// 是必选参数，且没有默认值
			if required == "1" && defaultVal == "" {
				if rule == "none" {
					check = false
				} else {
					return errors.New(name + " not found")
				}
			} else {
				check = false
			}
		}

		if check {
			switch rule {
			case "str":
				if !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "int":
				if !CheckStringDigit(defaultVal) || !CheckIntScope(defaultVal, min, max) {
					return errors.New(msg)
				}
			case "digit":
				if !CheckStringDigit(defaultVal) || !CheckIntScope(defaultVal, min, max) {
					return errors.New(msg)
				}
			case "digitString":
				if !CheckStringDigit(defaultVal) || !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "sDigit":
				if !CheckStringCommaDigit(defaultVal) || !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "sAlpha":
				if !CheckStringCommaAlpha(defaultVal) || !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "url":
				if !CheckUrl(defaultVal) {
					return errors.New(msg)
				}
			case "alnum":
				if !CheckStringAlnum(defaultVal) || !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "priv":
				if !isPriv(defaultVal) {
					return errors.New(msg)
				}
			case "dateTime":
				if !CheckDateTime(defaultVal) {
					return errors.New(msg)
				}
			case "date":
				if !CheckDate(defaultVal) {
					return errors.New(msg)
				}
			case "time":
				if !checkTime(defaultVal) {
					return errors.New(msg)
				}
			case "chn":
				if !CheckStringCHN(defaultVal) {
					return errors.New(msg)
				}
			case "module":
				if !CheckStringModule(defaultVal) || !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "float":
				if !CheckFloat(defaultVal) {
					return errors.New(msg)
				}
			case "vnphone":
				if !IsVietnamesePhone(defaultVal) {
					return errors.New(msg)
				}
			case "filter":
				if !CheckStringLength(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}

				defaultVal = FilterInjection(defaultVal)
			case "uname": //会员账号
				if !CheckUName(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			case "upwd": //会员密码
				if !CheckUPassword(defaultVal, int(min), int(max)) {
					return errors.New(msg)
				}
			default:
				break
			}
		}
	}

	return nil
}

func validate(val string) map[string]string {
	result := make(map[string]string)

	values := strings.Split(val, ",")
	for _, item := range values {
		parts := strings.SplitN(item, "=", 2) // 使用 SplitN 以限制分割次数
		if len(parts) != 2 {
			result[parts[0]] = "1"
		} else {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

func getValue(m map[string]string, key string) string {
	if value, exists := m[key]; exists {
		return value
	}
	return ""
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
