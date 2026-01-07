package ju

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
)

type JsConf struct {
	confFile string
	confVar  string
}

// NewJsConf 返回一个 JsConf 配置对象
//
// confFile 配置文件，如果这个值保持为空，则使用应用可执行文件目录下的 conf.js 文件
//
// confVar 配置文件中配置对应的变量，这个值保持空时，使用默认变量 conf
// noinspection GoUnusedExportedFunction
func NewJsConf(confFile, confVar string) *JsConf {
	if confFile == "" {
		confFile = "conf.js"
	}
	if confVar == "" {
		confVar = "conf"
	}
	return &JsConf{confFile: confFile, confVar: confVar}
}
func (jc *JsConf) Load(conf any) bool {
	data, err := os.ReadFile(jc.confFile)
	if LogFail(err) {
		return false
	}
	return jc.Parse(string(data), conf)
}
func (jc *JsConf) Parse(data string, conf any) bool {
	vm := goja.New()
	v, err := vm.RunScript(jc.confFile, string(data))
	if LogFail(err) {
		return false
	}

	//尝试读取脚本返回值，脚本是一个 json 对象时，适用此种情况
	obj, ok := v.Export().(map[string]interface{})
	if !ok {
		//尝试读取设置的全局变量
		v = vm.Get(jc.confVar)
		obj, ok = v.Export().(map[string]interface{})
		if !ok {
			LogRed(fmt.Sprintf("must has a variable name is '%s' as a object", jc.confVar))
			return false
		}
	}
	//把获取的 map 对象序列化为字符串，然后重新解析为 Go 的数据结构
	str, _ := json.Marshal(obj)
	_ = json.Unmarshal(str, conf)
	return true
}

func (jc *JsConf) Save(conf any) bool {
	data, _ := json.MarshalIndent(conf, "", "\t")
	js := "let conf;\nconf = " + string(data)
	return SaveFile(jc.confFile, []byte(js))
}
