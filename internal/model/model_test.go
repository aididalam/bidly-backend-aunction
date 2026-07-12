package model

import (
	"encoding/json"
	"testing"
)

func TestMoneyJSON(t *testing.T) {
	var m Money
	if err := json.Unmarshal([]byte(`123.45`), &m); err != nil || m != 12345 {
		t.Fatalf("unmarshal: %d %v", m, err)
	}
	data, _ := json.Marshal(m)
	if string(data) != "123.45" {
		t.Fatalf("marshal: %s", data)
	}
	for _, bad := range []string{`-1`, `1.234`, `"1.00"`, `null`} {
		if json.Unmarshal([]byte(bad), &m) == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}
