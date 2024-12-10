package ju

import (
	"encoding/hex"
	"strings"
)

func HexEncode(data []byte) string {
	return strings.ToUpper(hex.EncodeToString(data))
}
func HexDecode(h string) []byte {
	data, err := hex.DecodeString(h)
	CheckTrace(err, 1)
	return data
}
