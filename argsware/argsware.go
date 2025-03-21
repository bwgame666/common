package argsware

import (
	"fmt"
	"github.com/bwgame666/common/libs"
	"github.com/valyala/fasthttp"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// IsFormBody 判断请求体是否为 form 类型
func IsFormBody(ctx *fasthttp.RequestCtx) bool {
	contentType := string(ctx.Request.Header.ContentType())
	// 提取主类型部分
	mainContentType := strings.Split(contentType, ";")[0]
	// 判断是否为 form 类型
	return mainContentType == "application/x-www-form-urlencoded" ||
		mainContentType == "multipart/form-data"
}

// IsJSONBody 判断请求体是否为 JSON 类型
func IsJSONBody(ctx *fasthttp.RequestCtx) bool {
	contentType := string(ctx.Request.Header.ContentType())
	// 提取主类型部分
	mainContentType := strings.Split(contentType, ";")[0]
	// 判断是否为 JSON 类型
	return mainContentType == "application/json"
}

func JsonBindArgs(ctx *fasthttp.RequestCtx, structPointer interface{}) error {
	name := reflect.TypeOf(structPointer).String()
	api, err := GetParamsAPI(name)
	if err != nil {
		api, err = defaultArgsWare.register(structPointer)
		if err != nil {
			fmt.Println("register error:", err.Error())
			return err
		}
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewArgsError(api.name, "?", fmt.Sprint(p))
		}
	}()

	body := ctx.PostBody()
	if err = api.bodyJONS(reflect.ValueOf(structPointer), body); err != nil {
		return err
	}

	//验证required字段
	var paramsMap map[string]interface{}
	if err = libs.JsonUnmarshal(body, &paramsMap); err != nil {
		return err
	}

	fields := api.fieldsForBinding(reflect.ValueOf(structPointer).Elem())
	for i, param := range api.params {
		value := fields[i]

		_, found := getValueByPath(paramsMap, param.indexPath)
		if !found {
			if len(param.defaultVal) > 0 {
				if err = convertAssign(value, []string{param.defaultVal}); err != nil {
					return param.myError(err.Error())
				}
			} else if param.IsRequired() {
				return NewArgsError(param.apiName, param.FullName(), "is required.")
			}
		}

		if err = param.validate(value); err != nil {
			return err
		}
	}
	return nil
}

func QueryBindArgs(ctx *fasthttp.RequestCtx, structPointer interface{}) error {
	name := reflect.TypeOf(structPointer).String()
	api, err := GetParamsAPI(name)
	if err != nil {
		api, err = defaultArgsWare.register(structPointer)
		if err != nil {
			return err
		}
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewArgsError(api.name, "?", fmt.Sprint(p))
		}
	}()

	fields := api.fieldsForBinding(reflect.ValueOf(structPointer).Elem())
	for i, param := range api.params {
		value := fields[i]

		if paramValuesBytes := ctx.QueryArgs().PeekMulti(param.name); len(paramValuesBytes) > 0 {
			var paramValues = make([]string, len(paramValuesBytes))
			for i, b := range paramValuesBytes {
				paramValues[i] = string(b)
			}
			if err = convertAssign(value, paramValues); err != nil {
				return param.myError(err.Error())
			}
		} else if len(param.defaultVal) > 0 {
			if err = convertAssign(value, []string{param.defaultVal}); err != nil {
				return param.myError(err.Error())
			}
		} else if param.IsRequired() {
			return NewArgsError(param.apiName, param.FullName(), "is required.")
		}

		if err = param.validate(value); err != nil {
			return err
		}
	}
	return nil
}

func FormBindArgs(ctx *fasthttp.RequestCtx, structPointer interface{}) error {
	name := reflect.TypeOf(structPointer).String()
	api, err := GetParamsAPI(name)
	if err != nil {
		api, err = defaultArgsWare.register(structPointer)
		if err != nil {
			return err
		}
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewArgsError(api.name, "?", fmt.Sprint(p))
		}
	}()

	fields := api.fieldsForBinding(reflect.ValueOf(structPointer).Elem())
	var formValues = fasthttpFormValues(ctx)
	for i, param := range api.params {
		value := fields[i]

		if paramValues, ok := formValues[param.name]; ok {
			if err = convertAssign(value, paramValues); err != nil {
				return param.myError(err.Error())
			}
		} else if len(param.defaultVal) > 0 {
			if err = convertAssign(value, []string{param.defaultVal}); err != nil {
				return param.myError(err.Error())
			}
		} else if param.IsRequired() {
			return NewArgsError(param.apiName, param.FullName(), "is required.")
		}

		if err = param.validate(value); err != nil {
			return err
		}
	}
	return nil
}

func Validate(structPointer interface{}) error {
	name := reflect.TypeOf(structPointer).String()
	api, err := GetParamsAPI(name)
	if err != nil {
		api, err = defaultArgsWare.register(structPointer)
		if err != nil {
			return err
		}
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewArgsError(api.name, "?", fmt.Sprint(p))
		}
	}()

	fields := api.fieldsForBinding(reflect.ValueOf(structPointer).Elem())
	for i, param := range api.params {
		value := fields[i]

		if param.IsRequired() {
			return NewArgsError(param.apiName, param.name, "is required.")
		}

		if err = param.validate(value); err != nil {
			return err
		}
	}
	return nil
}

var (
	defaultArgsWare = &ArgsWare{
		lib: map[string]*ParamsAPI{},
		rule: map[string]RuleChecker{
			"regexp": checkReg,
		},
	}
)

type RuleChecker func(interface{}, string, string) error

type ArgsWare struct {
	lib  map[string]*ParamsAPI
	rule map[string]RuleChecker
	sync.RWMutex
}

func (c *ArgsWare) get(paramsAPIName string) (*ParamsAPI, bool) {
	c.RLock()
	defer c.RUnlock()
	m, ok := c.lib[paramsAPIName]
	return m, ok
}

func (c *ArgsWare) set(m *ParamsAPI) {
	c.Lock()
	c.lib[m.name] = m
	defer c.Unlock()
}

func (c *ArgsWare) register(structPointer interface{}) (*ParamsAPI, error) {
	name := reflect.TypeOf(structPointer).String()
	v := reflect.ValueOf(structPointer)
	if v.Kind() != reflect.Ptr {
		return nil, NewArgsError(name, "*", "the binding object must be a struct pointer")
	}
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return nil, NewArgsError(name, "*", "the binding object must be a struct pointer")
	}
	var m = &ParamsAPI{
		name:             name,
		params:           []*Param{},
		structType:       v.Type(),
		rawStructPointer: structPointer,
	}

	err := m.addFields([]IndexPath{}, m.structType, v)
	if err != nil {
		return nil, err
	}
	defaultArgsWare.set(m)
	return m, nil
}

func checkReg(fieldValue interface{}, fieldName, param string) error {
	value, ok := fieldValue.(string)
	if !ok {
		return NewArgsError("checkReg", fieldName, "the fieldValue must be a string")
	}

	matched, err := regexp.MatchString(param, value)
	if err != nil {
		return err
	}
	if !matched {
		return NewArgsError("checkReg", fieldName, fmt.Sprintf("%s not match regexp %s", value, param))
	}

	return nil
}

func Register(rule string, fn RuleChecker) {
	if _, ok := defaultArgsWare.rule[rule]; !ok {
		defaultArgsWare.rule[rule] = fn
	}
}
