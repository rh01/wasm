// +build js,wasm

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

//go:generate cp $GOROOT/misc/wasm/wasm_exec.js ../../assets/

// load local image and return encoded with base64
func loadImage(path string) string {
	href := js.Global().Get("location").Get("href")
	u, err := url.Parse(href.String())
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path
	u.RawQuery = fmt.Sprint(time.Now().UnixNano())
	log.Println("loading image file: ", u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func main() {
	fmt.Println("hello wasm go")
	// get canvas environment
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")
	// set canvas style 500*500
	canvas.Set("width", js.ValueOf(500))
	canvas.Set("height", js.ValueOf(500))

	// load image
	images := make([]js.Value, 3)
	files := []string{
		"/data/out01.png",
		"/data/out02.png",
		"/data/out03.png",
		// "/data/out02.png",
	}
	for i, file := range files {
		// generate <img > element
		images[i] = js.Global().Call("eval", "new Image()")
		// set attributes src to base64 encoded data
		images[i].Set("src", "data:image/png;base64,"+loadImage(file))
	}

	// set eventListener bind with image, when image clicked, alert
	canvas.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		js.Global().Get("window").Call("alert", "Don't click me")
		return nil
	}))

	// call js setInterval(callback, interval) method to display annimal
	n := 0
	js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ctx.Call("clearRect", 0, 0, 500, 500)
		ctx.Call("drawImage", images[n%3], 0, 0)
		n++

		style := canvas.Get("style")
		left := style.Get("left")
		if left.Equal(js.Undefined()) {
			left = js.ValueOf("0px")
		} else {
			n, _ := strconv.Atoi(strings.TrimRight(left.String(), "px"))
			left = js.ValueOf(fmt.Sprintf("%dpx", n+10))
		}
		style.Set("left", left)
		return nil
	}), js.ValueOf(50))
	select {}
}
