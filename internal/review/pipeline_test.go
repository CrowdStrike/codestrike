package review

import "testing"

func TestIsIgnoredPath(t *testing.T) {
	pipeline := &Pipeline{}
	patterns := []string{
		"vendor/",
		"*.lock",
		"web/*.min.js",
	}

	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "vendor/package/source.go", want: true},
		{filename: "nested/vendor/source.go", want: false},
		{filename: "go.sum", want: false},
		{filename: "frontend/package.lock", want: true},
		{filename: "web/app.min.js", want: true},
		{filename: "nested/web/app.min.js", want: false},
		{filename: "web/app.js", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := pipeline.isIgnoredPath(tt.filename, patterns); got != tt.want {
				t.Errorf("isIgnoredPath(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
