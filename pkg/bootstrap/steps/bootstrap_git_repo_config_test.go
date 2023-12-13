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
			name: "should error for ssh url without scheme",
			args: args{
				url:    "git@github.com/example/cli-dev",
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

func Test_normaliseUrl(t *testing.T) {
	tests := []struct {
		name              string
		url               string
		wantNormalisedUrl string
		wantScheme        string
		wantErr           assert.ErrorAssertionFunc
	}{
		{
			name:              "should normalise https url with .git",
			url:               "https://github.com/username/repository.git",
			wantNormalisedUrl: "https://github.com/username/repository.git",
			wantScheme:        httpsScheme,
			wantErr:           assert.NoError,
		},
		{
			name:              "should normalise https url",
			url:               "https://github.com/username/repository",
			wantNormalisedUrl: "https://github.com/username/repository",
			wantScheme:        httpsScheme,
			wantErr:           assert.NoError,
		},
		{
			name:              "should normalise ssh url without scheme",
			url:               "git@github.com:weaveworks/weave-gitops.git",
			wantNormalisedUrl: "ssh://git@github.com/weaveworks/weave-gitops.git",
			wantScheme:        sshScheme,
			wantErr:           assert.NoError,
		},
		{
			name:              "should normalise ssh url with scheme",
			url:               "ssh://git@github.com/username/repository.git",
			wantNormalisedUrl: "ssh://git@github.com/username/repository.git",
			wantScheme:        sshScheme,
			wantErr:           assert.NoError,
		},
		{
			name: "should fail for invalid url",
			url:  "invalid_url",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid repository scheme")
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNormalisedUrl, gotScheme, err := normaliseUrl(tt.url)
			if !tt.wantErr(t, err, fmt.Sprintf("normaliseUrl(%v)", tt.url)) {
				return
			}
			assert.Equalf(t, tt.wantNormalisedUrl, gotNormalisedUrl, "normaliseUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantScheme, gotScheme, "normaliseUrl(%v)", tt.url)
		})
	}
}
