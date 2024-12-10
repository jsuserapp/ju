package ju

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"runtime"
	"time"
)

// CheckTrace
// trace = 0 记录使用这个函数的位置，trace = 1 记录上一级调用位置
// noinspection GoUnusedExportedFunction
func CheckTrace(err error, skip int) bool {
	if err != nil {
		logToColor(skip+3, "red", "", err.Error())
		return true
	}
	return false
}

// noinspection GoUnusedExportedFunction
func CheckError(err error) {
	if err != nil {
		logToColor(3, "red", "", err.Error())
	}
}

// noinspection GoUnusedExportedFunction
func CheckSuccess(err error) bool {
	if err != nil {
		logToColor(3, "red", "", err.Error())
		return false
	}
	return true
}
func CheckFailure(err error) bool {
	if err != nil {
		logToColor(3, "red", "", err.Error())
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
func JsonParse(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	CheckTrace(err, 1)
}
func JsonByte(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
