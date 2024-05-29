package service

import (
	"fmt"
	"reflect"
)

type HttpService struct{}
type RequestHandler func(req interface{}) (resp interface{})

func GetReflect(handle RequestHandler) {

	// 获取函数类型
	funcType := reflect.TypeOf(handle)

	// 获取函数的输入类型
	inputTypes := make([]reflect.Type, funcType.NumIn())
	for i := 0; i < funcType.NumIn(); i++ {
		inputTypes[i] = funcType.In(i)
	}
	fmt.Println("Input Types:", inputTypes)

	// 获取函数的输出类型
	outputTypes := make([]reflect.Type, funcType.NumOut())
	for i := 0; i < funcType.NumOut(); i++ {
		outputTypes[i] = funcType.Out(i)
	}
	fmt.Println("Output Types:", outputTypes)
}

func (that *HttpService) post(path string, handle RequestHandler) {

}

func (that *HttpService) get(path string, handle RequestHandler) {

}

func (that *HttpService) put(path string, handle RequestHandler) {

}

func (that *HttpService) delete(path string, handle RequestHandler) {

}

func (that *HttpService) head(path string, handle RequestHandler) {

}

func (that *HttpService) options(path string, handle RequestHandler) {

}

func (that *HttpService) patch(path string, handle RequestHandler) {

}
