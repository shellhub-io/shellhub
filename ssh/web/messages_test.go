package web

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMessageMinSize(t *testing.T) {
	msg := Message{
		Kind: 0,
		Data: "",
	}
	j, _ := json.Marshal(msg)

	if len(j) != MessageMinSize {
		t.Errorf("expected %d, got %d", MessageMinSize, len(j))
	}
}

func ExampleMessage() {
	msg := Message{
		Kind: 0,
		Data: "",
	}
	j, _ := json.Marshal(msg)

	fmt.Println(string(j)) // Output: {"kind":0,"data":""}
}
