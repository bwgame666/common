package argsware

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
)

type (
	Param struct {
		apiName    string // ParamsAPI name
		name       string // param name
		indexPath  []IndexPath
		defaultVal string            // 缺省值
		isRequired bool              // file is required or not
		tags       map[string]string // struct tags for this param
		rawTag     reflect.StructTag // the raw tag
		rawValue   reflect.Value     // the raw tag value
		arg        string            // 自定义规则的参数
		err        error             // the custom error for binding or validating
	}

	IndexPath struct {
		Index int
		Name  string
	}
)

const accuracy = 0.0000001

// Raw gets the param's original value
func (param *Param) Raw() interface{} {
	return param.rawValue.Interface()
}

// APIName gets ParamsAPI name
func (param *Param) APIName() string {
	return param.apiName
}

// Name gets parameter field name
func (param *Param) Name() string {
	return param.name
}

// FullName gets parameter field name with indexPath
func (param *Param) FullName() string {
	if len(param.indexPath) < 2 {
		return param.name
	}
	var result string
	if len(param.indexPath) > 1 {
		for idx, indexPath := range param.indexPath {
			if idx == 0 {
				result = indexPath.Name
			} else {
				result += fmt.Sprintf(".%s", indexPath.Name)
			}
		}
	}
	return result
}

// IsRequired tests if the param is declared
func (param *Param) IsRequired() bool {
	return param.isRequired
}

func (param *Param) myError(reason string) error {
	if param.err != nil {
		return param.err
	}
	return NewArgsError(param.apiName, param.name, reason)
}

func (param *Param) validate(value reflect.Value) (err error) {
	defer func() {
		p := recover()
		if param.err != nil {
			if err != nil {
				err = param.err
			}
		} else if p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	if value.Kind() == reflect.Struct {
		return nil
	}

	strMin, _ := param.tags["min"]
	strMax, _ := param.tags["max"]

	switch value.Kind() {
	case reflect.Struct:
		return nil
	case reflect.Slice, reflect.Array:
		if err = param.validateLen(value.Len(), strMin, strMax, param.name); err != nil {
			return err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if err = param.validateInt(value.Int(), strMin, strMax, param.name); err != nil {
			return err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if err = param.validateInt(value.Int(), strMin, strMax, param.name); err != nil {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if err = param.validateFloat(value.Float(), strMin, strMax, param.name); err != nil {
			return err
		}
	case reflect.String:
		if err = param.validateLen(value.Len(), strMin, strMax, param.name); err != nil {
			return err
		}
	}

	if rule, ok := param.tags["rule"]; ok { // 自定义规则校验
		if fn, ok := defaultArgsWare.rule[rule]; ok {
			if err = fn(value.Interface(), param.FullName(), param.arg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (param *Param) validateLen(length int, min, max, paramName string) error {
	if len(min) > 0 {
		minInt, err := strconv.Atoi(min)
		if err != nil {
			panic(err)
		}
		if length < minInt {
			return NewValidationError(ValidationErrorValueTooShort, paramName)
		}
	}
	if len(max) > 0 {
		maxInt, err := strconv.Atoi(max)
		if err != nil {
			panic(err)
		}
		if length > maxInt {
			return NewValidationError(ValidationErrorValueTooLong, paramName)
		}
	}
	return nil
}

func (param *Param) validateFloat(f64 float64, min, max, paramName string) error {
	if len(min) > 0 {
		minFloat, err := strconv.ParseFloat(min, 64)
		if err != nil {
			return err
		}
		if math.Min(f64, minFloat) == f64 && math.Abs(f64-minFloat) > accuracy {
			return NewValidationError(ValidationErrorValueTooSmall, paramName)
		}
	}
	if len(max) > 0 {
		maxFloat, err := strconv.ParseFloat(max, 64)
		if err != nil {
			return err
		}
		if math.Max(f64, maxFloat) == f64 && math.Abs(f64-maxFloat) > accuracy {
			return NewValidationError(ValidationErrorValueTooBig, paramName)
		}
	}
	return nil
}

func (param *Param) validateInt(v int64, min, max, paramName string) error {
	if len(min) > 0 {
		minInt, err := strconv.ParseInt(min, 10, 64)
		if err != nil {
			return err
		}
		if v < minInt {
			return NewValidationError(ValidationErrorValueTooSmall, paramName)
		}
	}
	if len(max) > 0 {
		maxInt, err := strconv.ParseInt(max, 10, 64)
		if err != nil {
			return err
		}
		if v > maxInt {
			return NewValidationError(ValidationErrorValueTooBig, paramName)
		}
	}
	return nil
}

func (param *Param) validateRegexp(s, reg, paramName string) error {
	matched, err := regexp.MatchString(reg, s)
	if err != nil {
		return err
	}
	if !matched {
		return NewValidationError(ValidationErrorValueNotMatch, paramName)
	}
	return nil
}
