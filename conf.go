package ju

import (
	"encoding/json"
	"github.com/dop251/goja"
	"os"
)

func LoadJsConf(js string, conf any) bool {
	vm := goja.New()
	v, err := vm.RunScript("conf.js", js)
	if CheckFailure(err) {
		return false
	}

	//尝试读取脚本返回值，脚本是一个 json 对象时，适用此种情况
	obj, ok := v.Export().(map[string]interface{})
	if !ok {
		//尝试读取 conf 的全局变量
		v = vm.Get("conf")
		obj, ok = v.Export().(map[string]interface{})
		if !ok {
			LogRed("must has a variable name is 'conf' as a object")
			return false
		}
	}
	//把获取的 map 对象序列化为字符串，然后重新解析为 Go 的数据结构
	str, _ := json.Marshal(obj)
	_ = json.Unmarshal(str, conf)
	return true
}

// noinspection GoUnusedExportedFunction
func LoadJsConfFile(fn string, conf any) bool {
	data, err := os.ReadFile(fn)
	if CheckFailure(err) {
		return false
	}
	return LoadJsConf(string(data), conf)
}

// noinspection GoUnusedExportedFunction
func SaveJsConfFile(fn string, conf any) bool {
	data, _ := json.MarshalIndent(conf, "", "\t")
	js := "let conf = " + string(data)
	return SaveFile(fn, []byte(js))
}
