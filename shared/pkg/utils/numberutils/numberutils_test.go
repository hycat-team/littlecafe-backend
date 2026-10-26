package numberutils

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{name: "bytes", in: 12, want: "12 byte"},
		{name: "kib", in: 2048, want: "2 KB"},
		{name: "mib whole", in: 3 * 1024 * 1024, want: "3 MB"},
		{name: "mib fraction", in: 1536 * 1024, want: "1.5 MB"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatBytes(tc.in); got != tc.want {
				t.Fatalf("FormatBytes(%d) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
