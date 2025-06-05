package worker

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestResultMarshalJSON(t *testing.T) {
	cases := []struct {
		Recv Result
		Exp  string
	}{
		{Recv: Result{}, Exp: `"error":null`},
		{Recv: Result{Error: errors.New("test")}, Exp: `"error":"test"`},
	}

	for _, c := range cases {
		bytes, err := json.Marshal(c.Recv)
		if err != nil {
			t.Fatalf("failed to serialize Result: %v", err)
		}

		if !strings.Contains(string(bytes), c.Exp) {
			t.Errorf("expected result to contain '%s', but didn't find it: %s", c.Exp, string(bytes))
		}
	}
}

func TestResultUnmarshalJSON(t *testing.T) {
	errStr := "test"

	cases := []struct {
		Recv []byte
		Exp  *string
	}{
		{Recv: []byte(`{"error":null}`), Exp: nil},
		{Recv: []byte(`{"error":"test"}`), Exp: &errStr},
	}

	for _, c := range cases {
		var res Result
		if err := json.Unmarshal(c.Recv, &res); err != nil {
			t.Fatalf("failed to deserialize Result: %v", err)
		}

		if res.Error == nil && c.Exp != nil {
			t.Errorf("expected to find an error, but didn't")
		}

		if res.Error != nil && c.Exp == nil {
			t.Errorf("expected no errors, but found an error")
		}

		if res.Error != nil && c.Exp != nil && res.Error.Error() != *c.Exp {
			t.Errorf("the received error did not match the expected error")
		}
	}
}
