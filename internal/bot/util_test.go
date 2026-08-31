package bot

import (
	"context"
	"errors"
	"testing"
)

func TestValidMove(t *testing.T) {
	valid := []string{"F", "F'", "B", "B'", "U", "U'", "D", "D'", "L", "L'", "R", "R'", "x", "x'", "y", "y'", "z", "z'"}
	for _, m := range valid {
		if !validMove(m) {
			t.Fatalf("expected %q to be a valid move", m)
		}
	}
	invalid := []string{"", "Q", "X", "M", "F2", "R2", "u", "l", "b", "d", "f", "r"}
	for _, m := range invalid {
		if validMove(m) {
			t.Fatalf("expected %q to be an invalid move", m)
		}
	}
}

func TestDoWithRetryStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := 0
	err := doWithRetry(ctx, func() error {
		called++
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	if called != 1 {
		t.Fatalf("expected exactly one attempt once the context is cancelled, got %d", called)
	}
}

func TestDoWithRetryRetriesOnFailure(t *testing.T) {
	called := 0
	err := doWithRetry(context.Background(), func() error {
		called++
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected an error after retries are exhausted")
	}
	if called != defaultAttempts {
		t.Fatalf("expected %d attempts on persistent failure, got %d", defaultAttempts, called)
	}
}

func TestDoWithRetrySucceeds(t *testing.T) {
	called := 0
	err := doWithRetry(context.Background(), func() error {
		called++
		if called < 2 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if called != 2 {
		t.Fatalf("expected 2 attempts, got %d", called)
	}
}
