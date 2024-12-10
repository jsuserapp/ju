// main_test.go
package ju

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestPrint(t *testing.T) {
	js := JsonObject{}
	key := "person"
	js.SetValue(key, TestStruct{
		Name: "Jone",
		Age:  20,
	})
	person := &TestStruct{}
	rst := js.ReadValue(key, &person)
	LogRed(rst, person)

	key = "int"
	js.SetValue(key, 3)
	var i int
	rst = js.ReadValue(key, &i)
	LogGreen(rst, i)

	key = "array"
	js.SetValue(key, []float64{3, 6.2, 9})
	var a []int
	rst = js.ReadValue(key, &a)
	LogYellow(rst, a)
}
func TestDelay(t *testing.T) {
	usernames := []string{"alice", "bob", "carol"}
	for _, username := range usernames {
		go func() {
			time.Sleep(1 * time.Second)
			fmt.Println(username) // 这里可能不会按预期打印 "alice", "bob", "carol"
		}()
	}
	time.Sleep(1 * time.Second)

	for _, username := range usernames {
		un := username // 创建一个局部变量来捕获当前的值
		go func() {
			fmt.Println(un) // 这里将按预期打印 "alice", "bob", "carol"
		}()
	}
	time.Sleep(5 * time.Second)
}
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
