package main

import (
	"github.com/jsuserapp/ju"
)

func main() {
	serv := ju.NewLinuxService("ju", "", "")
	serv.Start(func() {
		ju.LogGreen("执行业务")
	})
}
