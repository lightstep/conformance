package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

func main() {
	argv := os.Args[1:]

	if len(argv) == 0 {
		panic(fmt.Sprint("no command provided", argv))
	}

	tracer := lightstep.NewTracer(lightstep.Options{
		AccessToken: "invalid",
	})

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = stdinReader
	cmd.Stdout = stdoutWriter
	cmd.Stderr = os.Stderr
	go func() {
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		log.Println("program has exited")
	}()

	ctx := newSpanContext(tracer)

	body, err := NewBodyFromContext(tracer, ctx)
	if err != nil {
		log.Fatal("couldnt create body ", err)
	}
	if err := json.NewEncoder(stdinWriter).Encode(body); err != nil {
		log.Fatal("could not marshall body: ", err)
	}

	if err := stdinWriter.Close(); err != nil {
		log.Println("failed to close writer")
	}

	var result Carriers
	if err := json.NewDecoder(stdoutReader).Decode(&result); err != nil {
		log.Fatal("could not decode ", err)
	}

	if err := result.Equals(tracer, ctx); err != nil {
		log.Println(body, result)
		log.Fatal(err)
	}

	log.Println("SUCCESS: span contexts are equal")
	cmd.Process.Kill()
	os.Exit(0)
}

func newSpanContext(tracer opentracing.Tracer) opentracing.SpanContext {
	ctx := tracer.StartSpan("fake").Context()
	lsctx, ok := ctx.(lightstep.SpanContext)
	if !ok {
		panic("could not convert context to lightstep context")
	}

	// Set IDs to uint64 max to test for any weirdness going over the int64 max.
	lsctx.SpanID = ^uint64(0)
	lsctx.TraceID = ^uint64(0)
	return lsctx
}
