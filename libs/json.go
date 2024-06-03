package libs

import (
	"github.com/bytedance/sonic"
)

func JsonMarshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}
