package main

import "testing"

func Test_connect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "test one"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connect()
		})
	}
}
