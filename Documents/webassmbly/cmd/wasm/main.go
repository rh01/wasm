package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
)

func main() {
	fmt.Println("Go Web assembly")
	js.Global().Set("formatJSON", jsonWrapper())
	<-make(chan struct{})
}

func prettyJSON(input string) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "\t")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

//expose Go function to jsfunction
func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(
		func(this js.Value, args []js.Value) interface{} {
			if len(args) != 1 {
				result := map[string]interface{}{
					"error": "Invalid no of arguments passed",
				}
				return result
			}

			jsonDoc := js.Global().Get("document")
			// 测试HTML元素是否存在，如果返回false，则表示该属性不存在
			if !jsonDoc.Truthy() {
				result := map[string]interface{}{
					"error": "Unable to get document object",
				}
				return result
			}
			// +translate js code：jsDoc.getElementById("jsonoutput")
			jsonOutputTextArea := jsonDoc.Call("getElementById", "jsonoutput")
			if !jsonOutputTextArea.Truthy() {
				result := map[string]interface{}{
					"error": "Unable to get output text area",
				}
				return result
			}
			inputJSON := args[0].String()
			fmt.Printf("input %s\n", inputJSON)
			pretty, err := prettyJSON(inputJSON)
			// 为了更好理解错误，需要建立从Golang到js的契约
			// 使用map
			if err != nil {
				errStr := fmt.Sprintf("unable to convert to json %s", err)
				result := map[string]interface{}{
					"error": errStr,
				}
				return result
			}
			// set the value property of the jsonoutput textarea
			jsonOutputTextArea.Set("value", pretty)
			return nil
		})
	return jsonFunc
}
