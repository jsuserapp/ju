package ju

import (
	"encoding/hex"
	"regexp"
	"strings"
)

func HexEncode(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}

// HexDecode 解码hex字符串，做了一些容错处理，自动移除空白字符
func HexDecode(h string) []byte {
	reg := regexp.MustCompile(`\s+`)
	h = reg.ReplaceAllString(h, "")
	if len(h) == 0 {
		return nil
	}
	data, err := hex.DecodeString(h)
	LogErrorTrace(err, 1)
	return data
}
