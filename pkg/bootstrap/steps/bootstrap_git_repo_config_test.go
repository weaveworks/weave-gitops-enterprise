package steps

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestNewGitRepositoryConfig(t *testing.T) {
	type args struct {
		url    string
		branch string
		path   string
	}
	tests := []struct {
		name    string
		args    args
		want    GitRepositoryConfig
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should create config for valid ssh url ",
			args: args{
				url:    "ssh://git@github.com/example/cli-dev",
				branch: "main",
				path:   "clusters/management",
			},
			want: GitRepositoryConfig{
				Url:    "ssh://git@github.com/example/cli-dev",
				Branch: "main",
				Path:   "clusters/management",
				Scheme: sshScheme,
			},
			wantErr: assert.NoError,
		},
		{
			name: "should create config for valid https url ",
			args: args{
				url:    "https://github.com/example/cli-dev",
				branch: "main",
				path:   "clusters/management",
			},
			want: GitRepositoryConfig{
				Url:    "https://github.com/example/cli-dev",
				Branch: "main",
				Path:   "clusters/management",
				Scheme: httpsScheme,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGitRepositoryConfig(tt.args.url, tt.args.branch, tt.args.path)
			if !tt.wantErr(t, err, fmt.Sprintf("NewGitRepositoryConfig(%v, %v, %v)", tt.args.url, tt.args.branch, tt.args.path)) {
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("unexpected git repository config:\n%s", diff)
			}
		})
	}
}
