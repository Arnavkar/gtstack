package main

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "''"},
		{"gt", "gt"},
		{"07-31-add_the_login_form", "07-31-add_the_login_form"},
		{"--no-trunk", "--no-trunk"},
		{"https://github.com/hSATAC/gtstack", "https://github.com/hSATAC/gtstack"},
		{"a b", "'a b'"},
		{"Add the login form", "'Add the login form'"},
		{"it's", `'it'\''s'`},
		{"$(rm -rf /)", `'$(rm -rf /)'`},
		{"a;b", "'a;b'"},
		// Non-ASCII is not in the alnum range, so it takes the quoted path.
		{"修正", "'修正'"},
	}
	for _, tt := range tests {
		if got := shellQuote(tt.in); got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBranchNameFrom(t *testing.T) {
	tests := []struct {
		msg, date, want string
	}{
		{"Add the login form", "07-31", "07-31-add_the_login_form"},
		{"CamelCase", "07-31", "07-31-camelcase"},
		{"v2 API", "07-31", "07-31-v2_api"},
		// Runs of punctuation collapse into a single separator.
		{"Fix   multiple   spaces", "01-02", "01-02-fix_multiple_spaces"},
		{"a---b", "01-02", "01-02-a_b"},
		// Leading separators never open the slug, trailing ones are trimmed.
		{"   leading", "01-02", "01-02-leading"},
		{"trailing!!!", "01-02", "01-02-trailing"},
		{"...both...", "01-02", "01-02-both"},
		// A message with nothing usable still has to yield a valid branch name.
		{"", "01-02", "01-02-branch"},
		{"!!!", "01-02", "01-02-branch"},
		// Non-ASCII drops out, but the ASCII around it survives.
		{"修正 bug", "01-02", "01-02-bug"},
	}
	for _, tt := range tests {
		if got := branchNameFrom(tt.msg, tt.date); got != tt.want {
			t.Errorf("branchNameFrom(%q, %q) = %q, want %q", tt.msg, tt.date, got, tt.want)
		}
	}
}

func TestJoinMessage(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"Subject"}, "Subject"},
		{[]string{"Subject", "Body"}, "Subject\n\nBody"},
		{[]string{"Subject", "Body", "More"}, "Subject\n\nBody\n\nMore"},
	}
	for _, tt := range tests {
		if got := joinMessage(tt.in); got != tt.want {
			t.Errorf("joinMessage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
