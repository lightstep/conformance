package main

import (
	"bytes"
	"encoding/base64"
	"fmt"

	lightstep "github.com/lightstep/lightstep-tracer-go"
	opentracing "github.com/opentracing/opentracing-go"
)

type Carriers struct {
	TextMap map[string]string `json:"text_map"`

	// Binary is the a base64 representation of the encoded byte array.
	Binary string `json:"binary"`
}

// NewBody creates a new request body and injects a span context into it.
func NewBodyFromContext(tracer opentracing.Tracer, ctx opentracing.SpanContext) (*Carriers, error) {
	body := &Carriers{
		TextMap: map[string]string{},
		Binary:  "",
	}

	err := tracer.Inject(ctx, opentracing.TextMap, opentracing.TextMapCarrier(body.TextMap))
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer(nil)
	if err := tracer.Inject(ctx, opentracing.Binary, buffer); err != nil {
		panic(err)
	}
	body.Binary = base64.StdEncoding.EncodeToString(buffer.Bytes())

	err = body.Equals(tracer, ctx)
	return body, err
}

func (b *Carriers) Equals(tracer opentracing.Tracer, ctx opentracing.SpanContext) error {
	original, ok := ctx.(lightstep.SpanContext)
	if !ok {
		panic("not lightstep context")
	}

	if err := b.checkTextMap(tracer, original); err != nil {
		return fmt.Errorf("text map: %v", err)
	}
	if err := b.checkBinary(tracer, original); err != nil {
		return fmt.Errorf("binary: %v", err)
	}
	return nil
}

func (b *Carriers) checkBinary(tracer opentracing.Tracer, original lightstep.SpanContext) error {
	bs, err := base64.StdEncoding.DecodeString(b.Binary)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(bs)
	ctx, err := tracer.Extract(opentracing.Binary, buf)
	if err != nil {
		return err
	}
	return contextsAreEqual(original, ctx)
}

func (b *Carriers) checkTextMap(tracer opentracing.Tracer, original lightstep.SpanContext) error {
	ctx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(b.TextMap))
	if err != nil {
		return err
	}
	return contextsAreEqual(original, ctx)
}

func contextsAreEqual(a lightstep.SpanContext, otcontext opentracing.SpanContext) error {
	if otcontext == nil {
		return fmt.Errorf("extracted context was nil")
	}

	b, ok := otcontext.(lightstep.SpanContext)
	if !ok {
		panic("not ok")
	}

	if a.TraceID != b.TraceID {
		return fmt.Errorf("expected %+v, got %+v", a, b)
	}
	if a.SpanID != b.SpanID {
		return fmt.Errorf("expected %+v, got %+v", a, b)
	}
	if len(a.Baggage) != len(b.Baggage) {
		return fmt.Errorf("expected %+v, got %+v", a, b)
	}

	for key, value := range a.Baggage {
		v, ok := b.Baggage[key]
		if !ok {
			return fmt.Errorf("extracted context does not have baggage for %v", key)
		}
		if v != value {
			return fmt.Errorf("expected value %s, got %s", value, v)
		}
	}
	return nil
}
