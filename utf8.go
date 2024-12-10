package ju

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
)

// GbkToUtf8 函数把 GBK 编码的字节数据转换为 UTF8 编码的字串，对于合法的 GBK 字节数据，这个函数不会失败，
// 但是如果源数据含有不能正常转换的数据，函数会反应一个错误和一个空字串。

func GbkToUtf8(s []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return "", e
	}
	return string(d), nil
}

// Utf8ToGbk 函数转换 Utf8 编码的字串(golang 字串内存编码)为 GBK 编码数据，
// 如果源字串 s 包含无法用 GBK 编码表示的字符，则函数会失败，此时返回的数据为一个长度为 0 的字节数组，但是不是 nil。
func Utf8ToGbk(s string) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewEncoder())
	return io.ReadAll(reader)
}

type UGSuccessProc func(gbk []byte, s string)
type UGFailedProc func(r rune, s string)

// UGEncode 和 Utf8ToGbk 功能相同，转换 utf8 为 gbk，但是这函数会逐字符转换，
// 成功调用 success 回调，失败调用 failed 回调。
func UGEncode(s string, success UGSuccessProc, failed UGFailedProc) {
	for _, r := range []rune(s) {
		ss := string(r)
		reader := transform.NewReader(bytes.NewReader([]byte(ss)), simplifiedchinese.GBK.NewEncoder())
		d, e := io.ReadAll(reader)
		if e != nil {
			failed(r, ss)
		} else {
			success(d, ss)
		}
	}
}
