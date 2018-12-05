package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	lightstep "github.com/lightstep/lightstep-tracer-go"
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
	go func() {
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	ctx := tracer.StartSpan("fake").Context()

	body, err := NewBodyFromContext(tracer, ctx)
	if err != nil {
		log.Fatalln("couldnt create body ", err)
	}
	if err := json.NewEncoder(stdinWriter).Encode(body); err != nil {
		log.Fatalln("could not marshall body: ", err)
	}

	var result Body
	if err := json.NewDecoder(stdoutReader).Decode(&result); err != nil {
		log.Fatal("could not decode ", err)
	}

	if err := result.Equals(tracer, ctx); err != nil {
		log.Println(body, results)
		log.Fatal(err)
	}

	log.Println("span contexts are equal")
	cmd.Process.Kill()
	os.Exit(0)
}
