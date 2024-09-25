package libs

import (
	"encoding/json"
	"fmt"
)

func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func JsonUnmarshalGeneral(data []byte) map[string]interface{} {
	var ret map[string]interface{}
	err := json.Unmarshal(data, &ret)
	if err != nil {
		fmt.Println("JsonUnmarshalGeneral failed: ", err)
		return nil
	}
	return ret
}
