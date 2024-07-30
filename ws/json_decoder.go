package ws

import "github.com/bwgame666/common/libs"

func JsonDecoder(p []byte) (interface{}, error) {
	ret := libs.JsonUnmarshalGeneral(p)
	return ret, nil
}

func JsonEncoder(data interface{}) ([]byte, error) {
	return libs.JsonMarshal(data)
}
