package main

import (
	"context"
	"fmt"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
)

func main() {
	// Initialize tracer
	shutdown := mmotel.InitTracer("test-service", mmotel.WithStdoutExporter())
	defer shutdown()

	// Create a context
	ctx := context.Background()

	// Test function that creates an error with stack trace
	result, err := testFunction(ctx)
	if err != nil {
		fmt.Println("Error occurred:", err)
		if mmErr, ok := err.(*mmerror.MmError); ok {
			fmt.Println("\nError with stack trace:")
			fmt.Println(mmErr.ErrorWithStack())
		}
	} else {
		fmt.Println("Result:", result)
	}
}

func testFunction(ctx context.Context) (string, error) {
	return nestedFunction(ctx)
}

func nestedFunction(ctx context.Context) (string, error) {
	return deeplyNestedFunction(ctx)
}

func deeplyNestedFunction(ctx context.Context) (string, error) {
	// Create an error with stack trace
	return "", mmerror.LogAndReturnError(
		ctx,
		mmerror.InternalServerError,
		"Something went wrong",
		fmt.Errorf("original error"),
		"Test error with stack trace",
	)
}
