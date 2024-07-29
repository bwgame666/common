package ws

import "github.com/bwgame666/common/libs"

func JsonDecoder(p []byte) (interface{}, error) {
	return p, nil
}

func JsonEncoder(data interface{}) ([]byte, error) {
	return libs.JsonMarshal(data)
}
