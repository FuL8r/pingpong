package pingpong_test

import (
	"testing"

	"github.com/Tuna/pingpong/internal/pingpong"
)

func TestService_Respond(t *testing.T) {
	tests := []struct {
		name string
		cmd  pingpong.Command
		want pingpong.Response
	}{
		{
			name: "valid NotBad header yields allowed ReallyNotBad",
			cmd:  pingpong.Command{NotBad: true},
			want: pingpong.Response{Decision: pingpong.Allowed, Body: "ReallyNotBad"},
		},
		{
			name: "missing NotBad header yields denied with empty body",
			cmd:  pingpong.Command{NotBad: false},
			want: pingpong.Response{Decision: pingpong.Denied, Body: ""},
		},
	}

	svc := pingpong.NewService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.Respond(tt.cmd)
			if got != tt.want {
				t.Fatalf("Respond(%+v) = %+v, want %+v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestService_Concurrent(t *testing.T) {
	svc := pingpong.NewService()
	const goroutines = 64

	done := make(chan struct{})
	for range goroutines {
		go func() {
			defer func() { done <- struct{}{} }()
			if got := svc.Respond(pingpong.Command{NotBad: true}); got.Body != "ReallyNotBad" {
				t.Errorf("unexpected body under concurrency: %q", got.Body)
			}
		}()
	}
	for range goroutines {
		<-done
	}
}
