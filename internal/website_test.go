package service

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteWebsite_InvalidID(t *testing.T) {
	service := NewWebsiteService(nil)

	err := service.DeleteWebsite(context.Background(), 0)
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("DeleteWebsite() error = %v, want %v", err, ErrInvalidID)
	}
}
