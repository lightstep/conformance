# LightStep Carrier Conformance

This repo contains a conformance runner that writes a span context to a client
library via stdout and reads the response from stdin.

# Carrier Specs

## OpenTracing TextMap
OpenTracing defines the TextMap as `a platform-idiomatic map from (unicode) string to string`
LightStep implements the TextMap by encoding the SpanContext into a map of string to string.
Given a LightStep SpanContext, we expect the following fields in the map

`ot-tracer-spanid` contains a uint encoded as base16 characters.
`ot-tracer-traceid` contains a uint encoded as bas16 characters.
`ot-tracer-sampled` is a boolean encoded as a string with the values `true` or `false`

Baggage items are key value pairs. All baggage keys are prefixed
with `ot-baggage-` and the corresponding value is the raw string.

## OpenTracing Binary


