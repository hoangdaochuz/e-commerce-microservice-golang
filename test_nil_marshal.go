package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// Test json.Marshal with nil
	b, _ := json.Marshal(nil)
	fmt.Printf("json.Marshal(nil): %v\n", b)
	fmt.Printf("string: %s\n", string(b))
	fmt.Printf("bytes: %v\n", []byte("null"))

	// Check if they match
	fmt.Printf("Match: %v\n", string(b) == "null")
}
