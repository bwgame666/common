package argsware

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"reflect"
	"strconv"
	"strings"
)

func fasthttpFormValues(req *fasthttp.RequestCtx) map[string][]string {
	// first check if we have multipart formValues
	multipartForm, err := req.MultipartForm()
	if err == nil {
		//we have multipart formValues
		return multipartForm.Value
	}
	valuesAll := make(map[string][]string)
	// if no multipart and post arguments ( means normal formValues   )
	if req.PostArgs().Len() == 0 {
		return valuesAll // no found
	}
	req.PostArgs().VisitAll(func(k []byte, v []byte) {
		key := string(k)
		value := string(v)
		// for slices
		if valuesAll[key] != nil {
			valuesAll[key] = append(valuesAll[key], value)
		} else {
			valuesAll[key] = []string{value}
		}
	})
	return valuesAll
}

func convertAssign(dest reflect.Value, src []string) (err error) {
	if len(src) == 0 {
		return nil
	}

	dest = reflect.Indirect(dest)
	if !dest.CanSet() {
		return fmt.Errorf("%s can not be setted", dest.Type().Name())
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	switch dest.Interface().(type) {
	case string:
		dest.Set(reflect.ValueOf(src[0]))
		return nil

	case []string:
		dest.Set(reflect.ValueOf(src))
		return nil

	case []byte:
		dest.Set(reflect.ValueOf([]byte(src[0])))
		return nil

	case [][]byte:
		b := make([][]byte, 0, len(src))
		for _, s := range src {
			b = append(b, []byte(s))
		}
		dest.Set(reflect.ValueOf(b))
		return nil

	case bool:
		dest.Set(reflect.ValueOf(parseBool(src[0])))
		return nil

	case []bool:
		b := make([]bool, 0, len(src))
		for _, s := range src {
			b = append(b, parseBool(s))
		}
		dest.Set(reflect.ValueOf(b))
		return nil
	}

	switch dest.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(src[0], 10, dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(src[0], 10, dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(src[0], dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetFloat(f64)
		return nil

	case reflect.Slice:
		member := dest.Type().Elem()
		switch member.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			for _, s := range src {
				i64, err := strconv.ParseInt(s, 10, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(i64).Convert(member)))
			}
			return nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for _, s := range src {
				u64, err := strconv.ParseUint(s, 10, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(u64).Convert(member)))
			}
			return nil

		case reflect.Float32, reflect.Float64:
			for _, s := range src {
				f64, err := strconv.ParseFloat(s, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(f64).Convert(member)))
			}
			return nil
		}
	}

	return fmt.Errorf("unsupported storing type %T into type %s", src, dest.Kind())
}

func parseBool(val string) bool {
	switch strings.TrimSpace(strings.ToLower(val)) {
	case "false", "off", "0":
		return false
	}
	return true
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

func getValueByPath(data map[string]interface{}, paths []IndexPath) (interface{}, bool) {
	var current interface{} = data

	for _, path := range paths {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current, ok = currentMap[path.Name]
			if !ok {
				return nil, false // 返回 nil 和 false 表示未找到
			}
		} else {
			return nil, false // 返回 nil 和 false 表示未找到
		}
	}

	return current, true // 返回找到的值和 true
}
