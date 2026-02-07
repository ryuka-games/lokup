package main

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Config
		wantErr bool
	}{
		{
			name: "basic repository",
			args: []string{"facebook/react"},
			want: &Config{
				Owner:  "facebook",
				Repo:   "react",
				Output: "report.html",
				Days:   30,
			},
		},
		{
			name: "with output flag",
			args: []string{"facebook/react", "--output", "custom.html"},
			want: &Config{
				Owner:  "facebook",
				Repo:   "react",
				Output: "custom.html",
				Days:   30,
			},
		},
		{
			name: "with days flag",
			args: []string{"facebook/react", "--days", "90"},
			want: &Config{
				Owner:  "facebook",
				Repo:   "react",
				Output: "report.html",
				Days:   90,
			},
		},
		{
			name: "with all flags",
			args: []string{"facebook/react", "--output", "out.html", "--days", "7"},
			want: &Config{
				Owner:  "facebook",
				Repo:   "react",
				Output: "out.html",
				Days:   7,
			},
		},
		{
			name:    "missing repository",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid repository format",
			args:    []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "empty owner",
			args:    []string{"/react"},
			wantErr: true,
		},
		{
			name:    "empty repo",
			args:    []string{"facebook/"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.Repo != tt.want.Repo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.want.Repo)
			}
			if got.Output != tt.want.Output {
				t.Errorf("Output = %q, want %q", got.Output, tt.want.Output)
			}
			if got.Days != tt.want.Days {
				t.Errorf("Days = %d, want %d", got.Days, tt.want.Days)
			}
		})
	}
}

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid",
			input:     "facebook/react",
			wantOwner: "facebook",
			wantRepo:  "react",
		},
		{
			name:      "with spaces",
			input:     " facebook / react ",
			wantOwner: "facebook",
			wantRepo:  "react",
		},
		{
			name:    "no slash",
			input:   "facebook",
			wantErr: true,
		},
		{
			name:    "too many slashes",
			input:   "facebook/react/extra",
			wantErr: true,
		},
		{
			name:    "empty owner",
			input:   "/react",
			wantErr: true,
		},
		{
			name:    "empty repo",
			input:   "facebook/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRepository(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}
