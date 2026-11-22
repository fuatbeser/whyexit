package main

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func itoa(n int) string { return strconv.Itoa(n) }

func runCLI(args ...string) (stdout, stderr string, status int) {
	var out, errb bytes.Buffer
	status = run(args, &out, &errb)
	return out.String(), errb.String(), status
}

func TestRun(t *testing.T) {
	tests := []struct {
		args   []string
		status int
		want   string
	}{
		{[]string{"127"}, 0, "127: command not found\n"},
		{[]string{"1"}, 0, "1: general error"},
		{[]string{"42"}, 0, "42: no standard meaning"},
		{[]string{"137"}, 0, "137 = 128 + 9 (SIGKILL): killed by force, often by the OOM killer\n"},
		{[]string{"130"}, 0, "130 = 128 + 2 (SIGINT): interrupted with Ctrl+C\n"},
		{[]string{"--os", "linux", "138"}, 0, "138 = 128 + 10 (SIGUSR1)"},
		{[]string{"--os", "macos", "138"}, 0, "138 = 128 + 10 (SIGBUS)"},
		{[]string{"--os", "windows", "1"}, 2, ""},
		{[]string{"200"}, 0, "200: no standard meaning"},
		{[]string{"abc"}, 2, ""},
		{[]string{}, 2, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, _, status := runCLI(tt.args...)
			if status != tt.status {
				t.Errorf("status = %d, want %d", status, tt.status)
			}
			if !strings.HasPrefix(out, tt.want) {
				t.Errorf("output = %q, want prefix %q", out, tt.want)
			}
		})
	}
}
