// +build js,wasm

package main

import (
	"syscall/js"
)

const (
	el  = "#app"
	msg = "welcome to wue.go"
)

func reverse(s string) string {
	var result string
	for _, v := range s {
		result = string(v) + result
	}
	return result
}

// Log function
func Log(i ...interface{}) {
	js.Global().Get("console").Call("log", i...)
}

// M is alias for map[string]interface{}
type M = map[string]interface{}

func main() {
	Vue := js.Global().Get("Vue")

	app := M{
		"el":   el,
		"data": M{"message": msg},
		"methods": M{
			// method1: reverse string
			"reverseMessage": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				data := js.Global().Get("app").Get("$data")
				if !data.Truthy() {
					result := map[string]interface{}{
						"error": "cannot find element $data",
					}
					return result
				}
				mess := reverse(data.Get("message").String())
				data.Set("message", mess)
				return nil
			}),
			// method2: call console.log print message
			"log": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				Log(js.Global().Get("app").Get("$data"))
				return nil
			})},
	}
	js.Global().Set("app", Vue.New(app))
	select {}
}
