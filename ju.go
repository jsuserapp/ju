package ju

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"runtime"
	"strconv"
	"time"
)

// CheckTrace
// trace = 0 记录使用这个函数的位置，trace = 1 记录上一级调用位置
// noinspection GoUnusedExportedFunction
func CheckTrace(err error, skip int) bool {
	if err != nil {
		logColor(skip+3, "red", "", err.Error())
		return true
	}
	return false
}

// noinspection GoUnusedExportedFunction
func CheckError(err error) {
	if err != nil {
		logColor(3, "red", "", err.Error())
	}
}

// noinspection GoUnusedExportedFunction
func CheckSuccess(err error) bool {
	if err != nil {
		logColor(3, "red", "", err.Error())
		return false
	}
	return true
}
func CheckFailure(err error) bool {
	if err != nil {
		logColor(3, "red", "", err.Error())
		return true
	}
	return false
}

func GetTrace(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	file = path.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}
func GetNowDateTime() string {
	return time.Now().Format(time.DateTime)
}
func GetNowDateTimeMs() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}
func GetNowTime() string {
	return time.Now().Format(time.TimeOnly)
}
func GetNowTimeMs() string {
	return time.Now().Format("15:04:05.000")
}

// noinspection GoUnusedExportedFunction
func GetNowDate() string {
	return time.Now().Format(time.DateOnly)
}

// noinspection GoUnusedExportedFunction
func RandString(length int, base string) string {
	if base == "" {
		base = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~@#$%^&*()_+-=,./<>?;:[]{}|"
	}
	var chars = []byte(base)
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("Error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue // Skip this number to avoid modulo bias.
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// noinspection GoUnusedExportedFunction
func Rand58String(n int) string {
	allowedChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	maxLen := len(allowedChars)
	b := ""
	for i := 0; i < n; i++ {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(maxLen)))
		pos := int(r.Int64())
		b += allowedChars[pos : pos+1]
	}
	return b
}
func JsonDecode(data []byte, v interface{}) bool {
	err := json.Unmarshal(data, v)
	CheckTrace(err, 1)
	return err == nil
}
func JsonDecodeString(str string, v interface{}) bool {
	err := json.Unmarshal([]byte(str), v)
	CheckTrace(err, 1)
	return err == nil
}
func JsonEncode(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
func JsonEncodeString(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
func StringBool(val string) bool {
	return val == "true"
}
func BoolString(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}
func StringFloat(val string) float64 {
	v, err := strconv.ParseFloat(val, 64)
	CheckTrace(err, 1)
	return v
}
func FloatString(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}
func StringInt(val string) int64 {
	v, err := strconv.ParseInt(val, 10, 64)
	CheckTrace(err, 1)
	return v
}
func IntString(val int64) string {
	return strconv.FormatInt(val, 10)
}
