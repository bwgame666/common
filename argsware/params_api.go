package argsware

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

type (
	ParamsAPI struct {
		name             string
		params           []*Param
		structType       reflect.Type
		rawStructPointer interface{}
	}
)

func GetParamsAPI(paramsAPIName string) (*ParamsAPI, error) {
	m, ok := defaultArgsWare.get(paramsAPIName)
	if !ok {
		return nil, errors.New("struct `" + paramsAPIName + "` is not registered")
	}
	return m, nil
}

func (m *ParamsAPI) parseTags(s string) map[string]string {
	c := strings.Split(s, ",")
	r := make(map[string]string)
	for _, v := range c {
		a := strings.IndexAny(v, "=:")
		if a != -1 {
			r[v[:a]] = v[a+1:]
			continue
		}
		r[v] = ""
	}
	return r
}

func (m *ParamsAPI) toSnake(s string) string {
	buf := bytes.NewBufferString("")
	for i, v := range s {
		if i > 0 && v >= 'A' && v <= 'Z' {
			buf.WriteRune('_')
		}
		buf.WriteRune(v)
	}
	return strings.ToLower(buf.String())
}

func (m *ParamsAPI) bodyJONS(dest reflect.Value, body []byte) error {
	var err error
	if dest.Kind() == reflect.Ptr {
		err = json.Unmarshal(body, dest.Interface())
	} else {
		err = json.Unmarshal(body, dest.Addr().Interface())
	}
	return err
}

func (m *ParamsAPI) addFields(parentIndexPath []IndexPath, t reflect.Type, v reflect.Value) error {
	var err error
	var deep = len(parentIndexPath) + 1
	for i := 0; i < t.NumField(); i++ {
		indexPath := make([]IndexPath, deep)
		copy(indexPath, parentIndexPath)
		indexPath[deep-1] = IndexPath{Name: m.toSnake(t.Field(i).Name), Index: i}

		var field = t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			path := indexPath
			if field.Anonymous {
				path = parentIndexPath
			}

			if err = m.addFields(path, field.Type, v.Field(i)); err != nil {
				return err
			}
			continue
		}

		tag, ok := field.Tag.Lookup("bind")
		if !ok {
			tag, ok = field.Tag.Lookup("validate")
		}
		if tag == "-" {
			continue
		}

		if field.Type.Kind() == reflect.Ptr {
			return NewArgsError(t.String(), field.Name, "field can not be a pointer")
		}

		var parsedTags = m.parseTags(tag)

		fd := &Param{
			apiName:   m.name,
			indexPath: indexPath,
			tags:      parsedTags,
			rawTag:    field.Tag,
			rawValue:  v.Field(i),
		}

		if errStr, ok := field.Tag.Lookup("msg"); ok {
			fd.tags["msg"] = errStr
			fd.err = errors.New(errStr)
		}
		if a, ok := field.Tag.Lookup("arg"); ok {
			fd.arg = a
		}
		if val, ok := parsedTags["default"]; ok {
			if (field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array) && len(val) > 0 {
				return NewArgsError(t.String(), field.Name, "invalid `default` tag for slice or array field")
			}
			fd.defaultVal = strings.TrimSpace(val)
		}

		if fd.name, ok = parsedTags["name"]; !ok {
			fd.name = m.toSnake(field.Name)
		}
		indexPath[deep-1] = IndexPath{Name: fd.name, Index: i}

		_, fd.isRequired = parsedTags["required"]

		m.params = append(m.params, fd)
	}
	return nil
}

func (m *ParamsAPI) fieldsForBinding(structElem reflect.Value) []reflect.Value {
	count := len(m.params)
	fields := make([]reflect.Value, count)
	for i := 0; i < count; i++ {
		value := structElem
		param := m.params[i]
		for _, indexPath := range param.indexPath {
			value = value.Field(indexPath.Index)
		}
		fields[i] = value
	}
	return fields
}
