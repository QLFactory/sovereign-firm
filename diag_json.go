package main

import (
	"encoding/json"
	"fmt"
)

type CodeBundle map[string]string

func main() {
	// Simulate LLM response with a nested object
	resp := `{"/package.json": {"name": "calc"}, "/App.js": "console.log('hi')"}`

	var rawBundle map[string]interface{}
	err := json.Unmarshal([]byte(resp), &rawBundle)
	if err != nil {
		fmt.Printf("Unmarshal into map[string]interface{} failed: %v\n", err)
		return
	}
	fmt.Println("Unmarshal into map[string]interface{} SUCCEEDED")

	bundle := make(CodeBundle)
	for k, v := range rawBundle {
		switch val := v.(type) {
		case string:
			bundle[k] = val
		default:
			marshaled, _ := json.MarshalIndent(val, "", "  ")
			bundle[k] = string(marshaled)
		}
	}

	fmt.Printf("Final Bundle: %+v\n", bundle)

	// Simulate Temporal unmarshaling the result
	marshaledResult, _ := json.Marshal(bundle)
	var finalBundle CodeBundle
	err = json.Unmarshal(marshaledResult, &finalBundle)
	if err != nil {
		fmt.Printf("Temporal-like unmarshal failed: %v\n", err)
	} else {
		fmt.Println("Temporal-like unmarshal SUCCEEDED")
	}
}
