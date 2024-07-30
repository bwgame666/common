package libs

import (
	"fmt"
	"github.com/bytedance/sonic"
)

func JsonMarshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

func JsonUnmarshalGeneral(data []byte) map[string]interface{} {
	var ret map[string]interface{}
	err := sonic.Unmarshal(data, &ret)
	if err != nil {
		fmt.Println("JsonUnmarshalGeneral failed: ", err)
		return nil
	}
	return ret
}
