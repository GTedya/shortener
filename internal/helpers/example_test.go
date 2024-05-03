package helpers

import "fmt"

// ExampleGenerateUUID demonstrates how to use the GenerateUUID function.
func ExampleGenerateUUID() {
	// Call GenerateUUID function with a file path
	uuid := GenerateUUID("example.txt")
	fmt.Println("Generated UUID:", uuid)
	// Output: Generated UUID: 1
}
