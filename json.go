package ju

import (
	"bytes"
	"encoding/json"
	"strconv"
)

/* Golang 的数组、结构等变量，实际上都是引用。
理论上数字字串也是引用，但是它们本身是不可改变的，所以是引用还是复制是没有区别的。
*/

type JsonObject map[string]any

func (js *JsonObject) String() string {
	b, err := json.Marshal(js)
	if err != nil {
		LogRed(err.Error())
		return ""
	}
	return string(b)
}
func (js *JsonObject) ToBytes() []byte {
	b, err := json.Marshal(js)
	if err != nil {
		LogRed(err.Error())
	}
	return b
}

// ParseAnyString 如果 str 不是用大括号括起来的 Object 字串，ParseString 就会失败，
// ParseAnyString 会把字串作为一个 key 来分析，这样只要是合法的 Json
// 字串，就能够返回正确的对象，比如布尔、字串，数字等等。
func (js *JsonObject) ParseAnyString(str string) (any, bool) {
	s := `{"a":` + str + `}`
	if nil != js.ParseString(s) {
		return nil, false
	}
	j := js.GetInterface("a")
	return j, true
}
func (js *JsonObject) ParseString(str string) error {
	return json.Unmarshal([]byte(str), js)
}
func (js *JsonObject) ToStruct(st any) error {
	data := js.ToBytes()
	return json.Unmarshal(data, st)
}
func (js *JsonObject) ParseBytes(b []byte) bool {
	err := json.Unmarshal(b, js)
	return err == nil
}
func (js *JsonObject) GetInterface(key string) any {
	return (*js)[key]
}
func (js *JsonObject) GetString(key string) (string, bool) {
	val := (*js)[key]
	switch val.(type) {
	case string:
		return val.(string), true
	default:
		return "", false
	}
}
func (js *JsonObject) GetUint64(key string) (uint64, bool) {
	val := (*js)[key]
	switch v := val.(type) {
	case uint64:
		return v, true
	case uint:
		return uint64(v), true
	}
	return 0, false
}
func (js *JsonObject) GetInt64(key string) (int64, bool) {
	val := (*js)[key]
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	}
	return 0, false
}

// GetNumber 如果这个键值是一个 Go 的数字数字类型，则返回 true，否则返回 false
func (js *JsonObject) GetNumber(key string) (float64, bool) {
	val := (*js)[key]
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	default:
		return 0, false
	}
}
func (js *JsonObject) GetArray(key string) ([]any, bool) {
	val := (*js)[key]
	switch val.(type) {
	case []any:
		return val.([]any), true
	default:
		return []any{}, false
	}
}
func (js *JsonObject) GetJson(key string) (JsonObject, bool) {
	v := (*js)[key]
	switch v.(type) {
	case map[string]any:
		jn := v.(map[string]any)
		return jn, true
	default:
		return nil, false
	}
}
func (js *JsonObject) GetBool(key string) (bool, bool) {
	val := (*js)[key]
	switch val.(type) {
	case bool:
		return val.(bool), true
	default:
		return false, false
	}
}

// ReadValue 读取值为特定类型，如果对应的 key 的值不能转换为 v 的类型，返回 false。因为 json 的数据类型
// 不一定和对应语言（这里是 Go）的数据类型完全匹配，比如 Go 的数字类型有很多种，但是 json 只有 float 一种，
// 为了避免用户自己判断和转换的麻烦，这里可以直接指定需要的数据类型，如果对应 key 的值可以被转换为这个类型，
// 则会自动转换，并且返回 true。
// v 必须是一个指针类型
// 某些情况下，函数返回 false，但是可能部分转换是成功的，这是因为 json.Unmarshal 函数的特性决定的，err 不为
// nil 时，也可能部分转换成功
func (js *JsonObject) ReadValue(key string, v any) bool {
	val, ok := (*js)[key]
	if !ok {
		return false
	}
	data, _ := json.Marshal(val)
	err := json.Unmarshal(data, v)
	return err == nil
}

// SetValue 这个函数会设置 map 对应的 key 值，但是可能和读取时是不一样的，因为反序列化 json 字串的对象都是 map，而设置的不一定，
// 你可以设置一个其它对象，它会被 json 序列化，而反序列化并不能还原这个对象，要反序列化为特定的对象使用 ReadValue 函数
func (js *JsonObject) SetValue(key string, v any) {
	(*js)[key] = v
}

// SetStrInt64 因为 json 的数字类型是 float64，和 Int64 的范围是不一样的，直接传递 Int64 类型的数据，有可能会不正确，这里用字符串来传递
func (js *JsonObject) SetStrInt64(key string, v int64) {
	vs := strconv.FormatInt(v, 10)
	(*js)[key] = vs
}
func (js *JsonObject) Delete(key string) {
	delete(*js, key)
}
func (js *JsonObject) Clear() {
	for key := range *js {
		delete(*js, key)
	}
}
func (js *JsonObject) Marshal(v any) {
	b, err := json.Marshal(v)
	if LogFail(err) {
		return
	}
	_ = js.ParseBytes(b)
}
func (js *JsonObject) UnMarshal(st any) bool {
	data := js.ToBytes()
	err := json.Unmarshal(data, st)
	if err != nil {
		OutputColor(1, ColorRed, err.Error())
	}
	return err == nil
}

// noinspection GoUnusedExportedFunction
func NewJsonObject() JsonObject {
	return JsonObject{}
}

type JsonArray []any

func (ja *JsonArray) ParseString(str string) error {
	return json.Unmarshal([]byte(str), ja)
}

func (ja *JsonArray) ParseBytes(b []byte) error {
	return json.Unmarshal(b, ja)
}
func (ja *JsonArray) UnMarshal(st any) bool {
	data := ja.ToBytes()
	err := json.Unmarshal(data, st)
	if err != nil {
		OutputColor(1, ColorRed, err.Error())
	}
	return err == nil
}
func (ja *JsonArray) Marshal(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		LogRed(err.Error())
		return
	}
	_ = ja.ParseBytes(b)
}
func (ja *JsonArray) ToString() string {
	b, err := json.Marshal(ja)
	if err != nil {
		LogRed(err.Error())
		return ""
	}
	return string(b)
}
func (ja *JsonArray) ToBytes() []byte {
	b, err := json.Marshal(ja)
	if err != nil {
		LogRed(err.Error())
	}
	return b
}

// noinspection GoUnusedExportedFunction
func NewJsonArray() JsonArray {
	return JsonArray{}
}

// noinspection GoUnusedExportedFunction
func JsonParseString(src string, pobj any) bool {
	err := json.Unmarshal([]byte(src), pobj)
	return err == nil
}
func JsonParseBytes(src []byte, pobj any) bool {
	err := json.Unmarshal(src, pobj)
	return err == nil
}

// noinspection GoUnusedExportedFunction
func JsonToString(obj any, format bool) string {
	data, _ := json.Marshal(obj)
	if format {
		var str bytes.Buffer
		err := json.Indent(&str, data, "", "\t")
		if err != nil {
			return ""
		}
		return str.String()
	} else {
		return string(data)
	}
}
func JsonToBytes(obj any, format bool) []byte {
	data, _ := json.Marshal(obj)
	if format {
		var str bytes.Buffer
		err := json.Indent(&str, data, "", "\t")
		if err != nil {
			return nil
		}
		return str.Bytes()
	} else {
		return data
	}
}

// noinspection GoUnusedExportedFunction
func JsonLoadConf(fn string, conf any) bool {
	data := ReadFile(fn)
	return JsonParseBytes(data, conf)
}

// noinspection GoUnusedExportedFunction
func JsonSaveConf(fn string, conf any) bool {
	data := JsonToBytes(conf, true)
	return SaveFile(fn, data)
}
