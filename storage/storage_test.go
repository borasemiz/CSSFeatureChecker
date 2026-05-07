package storage

import (
	"context"
	"testing"
)

func TestGetStorage_MissingBucket(t *testing.T) {
	t.Setenv("BUCKET_NAME", "")
	_, err := GetStorage(context.Background())
	if err == nil {
		t.Fatal("expected error when BUCKET_NAME is not set")
	}
}
