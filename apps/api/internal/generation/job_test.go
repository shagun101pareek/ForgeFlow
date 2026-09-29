package generation

import (
	"errors"
	"testing"
)

func TestFailureMessageHidesProviderDetails(t *testing.T) {
	if got := failureMessage(&ModelError{Message: "model overloaded"}); got != "model overloaded" {
		t.Fatalf("model message = %q", got)
	}
	if got := failureMessage(errors.New("openai api key is not configured")); got != "OpenAI API key is not configured" {
		t.Fatalf("key message = %q", got)
	}
	if got := failureMessage(errors.New("dial tcp: connection refused")); got != "could not generate a UI specification" {
		t.Fatalf("generic message = %q", got)
	}
}
