package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	byts, err := ioutil.ReadFile("fonts/luxisr.ttf")
	panicIfErr(err)
	var res bytes.Buffer
	res.WriteString("package main\n\n")
	res.WriteString(`var LuxiFont = []byte{`)
	for _, b := range byts {
		res.WriteString(fmt.Sprintf("%d", b))
		res.WriteString(",")
	}
	res.WriteString("}")
	panicIfErr(ioutil.WriteFile("cmd/gpxchart/font_generated.go", res.Bytes(), 0700))
}
