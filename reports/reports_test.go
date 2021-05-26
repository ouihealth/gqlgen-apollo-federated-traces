package reports

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestUnmarshalTrace(t *testing.T) {
	const encodedTrace = `GgsI15W0hQYQgJrJRiILCNeVtIUGEID/2kNYseWWA3KxAWKuAQoFYm9va3MaBltCb29rXUDT5OYCSKyo7QJiRRAAYh8KBXRpdGxlGgZTdHJpbmdA2oTzAkix+/UCagRCb29rYiAKBmF1dGhvchoGU3RyaW5nQPbl9wJI36H4AmoEQm9va2JFEAFiHwoFdGl0bGUaBlN0cmluZ0C4u/oCSJ23+wJqBEJvb2tiIAoGYXV0aG9yGgZTdHJpbmdA992AA0iSnoEDagRCb29ragVRdWVyeQ==`
	decodedTrace, err := base64.StdEncoding.DecodeString(encodedTrace)
	if err != nil {
		t.Fatal(err)
	}

	trace := &Trace{}
	if err := proto.Unmarshal(decodedTrace, trace); err != nil {
		t.Fatal(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	if err := e.Encode(trace); err != nil {
		t.Fatal(err)
	}
}
