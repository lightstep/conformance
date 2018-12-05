package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

const serverAddress = "http://127.0.0.1:6001"

type Body struct {
	TextMap map[string]string `json:"text_map"`

	Binary string `json:"binary"`
}

func main() {
	opentracing.SetGlobalTracer(lightstep.NewTracer(lightstep.Options{
		AccessToken: "invalid",
	}))
	tracer := opentracing.GlobalTracer()

	var body Body
	err := json.NewDecoder(os.Stdin).Decode(&body)
	if err != nil {
		panic(err)
	}
	spanContextTextMap, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(body.TextMap))
	if err != nil {
		panic(err)
	}

	b, err := base64.StdEncoding.DecodeString(body.Binary)
	if err != nil {
		panic(err)
	}

	spanContextBinary, err := tracer.Extract(opentracing.Binary, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	buffer2 := bytes.NewBuffer(nil)
	if err := tracer.Inject(spanContextBinary, opentracing.Binary, buffer2); err != nil {
		panic(err)
	}
	echoBody := Body{
		TextMap: make(map[string]string),
		Binary:  base64.StdEncoding.EncodeToString(buffer2.Bytes()),
	}

	err = tracer.Inject(spanContextTextMap, opentracing.TextMap, opentracing.TextMapCarrier(echoBody.TextMap))
	if err != nil {
		panic(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(echoBody); err != nil {
		panic(err)
	}
	return
}
