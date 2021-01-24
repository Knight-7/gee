package binding

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

const defaultMemory = 32 << 20

type formBinding struct{}
type formPostBinding struct{}
type multipartBinding struct{}

func (b formBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	return nil
}

func (b formPostBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	if err := mapForm(obj, req.PostForm); err != nil {
		return err
	}
	return nil
}

func (b multipartBinding) Bind(req *http.Request, obj interface{}) error {
	return nil
}

func mapForm(obj interface{}, form map[string][]string) error {
	_, err := mapping(reflect.ValueOf(obj), reflect.StructField{}, form, "form")
	return err
}

func mapping(value reflect.Value, field reflect.StructField, form map[string][]string, tag string) (bool, error) {
	// 当 tag 的值为 "-" 时，表示该字段不用解析
	if field.Tag.Get(tag) == "-" {
		return false, nil
	}

	valueKind := value.Kind()

	// value 是指针，判断指针是否为 nil； 当为 nil 是，要为其分配内存，
	// 否则通过 Elem 获取其指向的对象，递归调用该函数
	if valueKind == reflect.Ptr {
		isNil := false
		vPtr := value
		if value.IsNil() {
			isNil = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, form, tag)
		if err != nil {
			return false, err
		}
		if isNil && isSet {
			value.Set(vPtr)
		}
		return true, nil
	}

	// 当 value 不是结构体时且该字段可导出时，尝试给这个 value 赋值
	if valueKind != reflect.Struct || !field.Anonymous {
		ok, err := trySetValue(value, field, form, tag)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	// 当 value 是结构体时，遍历其所以可导出字段，递归调用该函数
	if valueKind == reflect.Struct {
		valueType := value.Type()

		isSet := false
		for i := 0; i < valueType.NumField(); i++ {
			vf := valueType.Field(i)
			// FieldStruct 中可导出字段的 PkgPath 是""
			if vf.PkgPath != "" && !vf.Anonymous {
				continue
			}
			ok, err := mapping(value.Field(i), vf, form, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}
	return false, nil
}

func trySetValue(value reflect.Value, field reflect.StructField, form map[string][]string, tag string) (bool, error) {
	tagValue := field.Tag.Get(tag)
	if tagValue == "" {
		tagValue = field.Name
	}
	if tagValue == "" {
		return false, nil
	}

	return setValue(value, tagValue, form, field, tag)
}

func setValue(value reflect.Value, tagValue string, form map[string][]string, field reflect.StructField, tag string) (bool, error) {
	v, ok := form[tagValue]
	if !ok {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		err := setSlice(v, value, field)
		if err != nil {
			return false, err
		}
		return true, nil
	case reflect.Array:
		err := setArray(v, value, field)
		if err != nil {
			return false, err
		}
		return true, nil
	default:
		return true, setProperValue(v[0], value, field)
	}
}

func setProperValue(val string, value reflect.Value, field reflect.StructField) error {
	switch value.Kind() {
	case reflect.Int:
		return setInt(val, 0, value)
	case reflect.Int8:
		return setInt(val, 8, value)
	case reflect.Int16:
		return setInt(val, 16, value)
	case reflect.Int32:
		return setInt(val, 32, value)
	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			return setDuration(val, value)
		}
		return setInt(val, 64, value)
	case reflect.Uint:
		return setUint(val, 0, value)
	case reflect.Uint8:
		return setUint(val, 8, value)
	case reflect.Uint16:
		return setUint(val, 16, value)
	case reflect.Uint32:
		return setUint(val, 32, value)
	case reflect.Uint64:
		return setUint(val, 64, value)
	case reflect.Float32:
		return setFloat(val, 32, value)
	case reflect.Float64:
		return setFloat(val, 64, value)
	case reflect.Bool:
		return setBool(val, value)
	case reflect.String:
		return setString(val, value)
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return setTime(val, value, field)
		}

		// FIXME: 字符串和字节数组的转换可能存在问题
		return json.Unmarshal([]byte(val), value.Addr().Interface())
	case reflect.Map:
		// FIXME: 字符串和字节数组的转化可能存在问题
		return json.Unmarshal([]byte(val), value.Addr().Interface())
	default:
		return errors.New("unknown type")
	}
}

func setArray(vals []string, value reflect.Value, field reflect.StructField) error {
	for i, val := range vals {
		err := setProperValue(val, value.Index(i), field)
		if err != nil {
			return err
		}
	}
	return nil
}

func setSlice(vals []string, value reflect.Value, field reflect.StructField) error {
	sliceValue := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, sliceValue, field)
	if err != nil {
		return err
	}
	value.Set(sliceValue)
	return nil
}

func setInt(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0"
	}
	parseInt, err := strconv.ParseInt(val, 10, bitSize)
	if err != nil {
		return err
	}
	value.SetInt(parseInt)
	return nil
}

func setUint(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0"
	}
	parseUint, err := strconv.ParseUint(val, 10, bitSize)
	if err != nil {
		return err
	}
	value.SetUint(parseUint)
	return nil
}

func setFloat(val string, bitSize int, value reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	parseFloat, err := strconv.ParseFloat(val, bitSize)
	if err != nil {
		return err
	}
	value.SetFloat(parseFloat)
	return nil
}

func setBool(val string, value reflect.Value) error {
	parseBool, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	value.SetBool(parseBool)
	return nil
}

func setString(val string, value reflect.Value) error {
	value.SetString(val)
	return nil
}

// TODO: 了解 time 包
func setTime(val string, value reflect.Value, field reflect.StructField) error {
	timeFormat := field.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(field.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := field.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	locTime, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(locTime))
	return nil
}

func setDuration(val string, value reflect.Value) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}
