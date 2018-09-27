package util

import (
	"encoding/json"
	"testing"
)

func TestGetMyInfo(t *testing.T) {
	my := GetMyInfo()

	b, _ := json.Marshal(my)

	t.Log(string(b))
}
