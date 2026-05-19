package main

import (
	"path/filepath"
	"testing"
)

func TestRepoFiltering(t *testing.T) {
	cfg := &Config{
		WorkRepos: []string{
			"/home/user/projects/platform-ops",
			"/home/user/projects/main-project",
		},
		PrivateRepos: []string{
			"/home/user/projects/personal-notes",
			"/home/user/projects/dotfiles",
		},
	}

	tests := []struct {
		name           string
		repoFullName   string
		includePrivate bool
		wantFiltered   bool
	}{
		{
			name:           "Work repo, work only mode",
			repoFullName:   "myorg/platform-ops",
			includePrivate: false,
			wantFiltered:   false,
		},
		{
			name:           "Private repo, work only mode",
			repoFullName:   "myuser/personal-notes",
			includePrivate: false,
			wantFiltered:   true,
		},
		{
			name:           "Private repo, all repos mode",
			repoFullName:   "myuser/personal-notes",
			includePrivate: true,
			wantFiltered:   false,
		},
		{
			name:           "Work repo, all repos mode",
			repoFullName:   "myorg/main-project",
			includePrivate: true,
			wantFiltered:   false,
		},
		{
			name:           "Dotfiles private repo, work only mode",
			repoFullName:   "myuser/dotfiles",
			includePrivate: false,
			wantFiltered:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateRepoNames := make(map[string]bool)
			if !tt.includePrivate {
				for _, repoPath := range cfg.PrivateRepos {
					repoName := filepath.Base(repoPath)
					privateRepoNames[repoName] = true
				}
			}

			parts := splitRepoFullName(tt.repoFullName)
			if len(parts) != 2 {
				t.Fatalf("Invalid repo name: %s", tt.repoFullName)
			}
			repoName := parts[1]

			shouldFilter := !tt.includePrivate && privateRepoNames[repoName]

			if shouldFilter != tt.wantFiltered {
				t.Errorf("Repo %s (includePrivate=%v): got filtered=%v, want %v",
					tt.repoFullName, tt.includePrivate, shouldFilter, tt.wantFiltered)
			}
		})
	}
}

func splitRepoFullName(fullName string) []string {
	result := []string{}
	current := ""
	for _, char := range fullName {
		if char == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func TestSplitRepoFullName(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		want     []string
	}{
		{
			name:     "Standard org/repo",
			fullName: "myorg/platform-ops",
			want:     []string{"myorg", "platform-ops"},
		},
		{
			name:     "User repo",
			fullName: "E40057167/SecondBrain",
			want:     []string{"E40057167", "SecondBrain"},
		},
		{
			name:     "Repo with hyphens",
			fullName: "org-name/repo-name-with-dashes",
			want:     []string{"org-name", "repo-name-with-dashes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitRepoFullName(tt.fullName)
			if len(got) != len(tt.want) {
				t.Fatalf("splitRepoFullName(%s) = %v, want %v", tt.fullName, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitRepoFullName(%s)[%d] = %s, want %s",
						tt.fullName, i, got[i], tt.want[i])
				}
			}
		})
	}
}
