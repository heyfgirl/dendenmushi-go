package main

import (
	"fmt"

	dendenmushi "github.com/heyfgirl/dendenmushi-go"
)

type body struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func main() {
	c, err := dendenmushi.NewClient("tcp", "127.0.0.1:8888", nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	body2 := &body{
		Method: "contact.update",
		Params: map[string]interface{}{"account_name": "yousri", "cellphone": "15581502447"},
	}
	var reply map[string]interface{}
	err = c.Call(body2.Method, body2.Params, &reply)
	fmt.Println(1111, reply)
}
