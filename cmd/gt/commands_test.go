package main

import (
	"slices"
	"testing"
)

func TestSubmitArgs(t *testing.T) {
	tests := []struct {
		name          string
		edit, publish bool
		want          []string
	}{
		// The bare `gt submit` must not open the editor: Graphite's does not,
		// and matching that is the point of the command.
		{"default", false, false, []string{"stack", "submit", "--auto"}},
		{"publish", false, true, []string{"stack", "submit", "--auto", "--open"}},
		{"edit", true, false, []string{"stack", "submit"}},
		{"edit and publish", true, true, []string{"stack", "submit", "--open"}},
	}
	for _, tt := range tests {
		if got := submitArgs(tt.edit, tt.publish); !slices.Equal(got, tt.want) {
			t.Errorf("%s: submitArgs(%v, %v) = %q, want %q", tt.name, tt.edit, tt.publish, got, tt.want)
		}
	}
}
