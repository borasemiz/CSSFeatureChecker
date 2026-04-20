package main

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestHandler(t *testing.T) {
	godotenv.Load()
	err := handler(context.Background())

	if err != nil {
		t.Fatalf("error happened %s", err.Error())
	}
}
