package web

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageMinSize(t *testing.T) {
	msg := Message{
		Kind: 0,
		Data: "",
	}
	j, err := json.Marshal(msg)
	require.NoError(t, err)

	if len(j) != MessageMinSize {
		t.Errorf("expected %d, got %d", MessageMinSize, len(j))
	}
}

func ExampleMessage() {
	msg := Message{
		Kind: 0,
		Data: "",
	}
	j, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j)) // Output: {"kind":0,"data":""}
}
