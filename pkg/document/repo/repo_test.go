package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-billy.v4/memfs"
	fixtures "gopkg.in/src-d/go-git-fixtures.v3"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"opendev.org/airship/airshipctl/testutil"
)

type mockBuilder struct {
	CloneOptions    *git.CloneOptions
	AuthMethod      transport.AuthMethod
	CheckoutOptions *git.CheckoutOptions
	FetchOptions    *git.FetchOptions
	URLString       string
	AuthError       error
}

func (md mockBuilder) ToAuth() (transport.AuthMethod, error) {
	return md.AuthMethod, md.AuthError
}
func (md mockBuilder) ToCloneOptions(transport.AuthMethod) *git.CloneOptions {
	return md.CloneOptions
}
func (md mockBuilder) ToCheckoutOptions(bool) *git.CheckoutOptions {
	return md.CheckoutOptions
}
func (md mockBuilder) ToFetchOptions(transport.AuthMethod) *git.FetchOptions {
	return md.FetchOptions
}
func (md mockBuilder) URL() string { return md.URLString }

func TestNewRepositoryNegative(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	defer testutil.CleanUpGitFixtures(t)

	builder := &mockBuilder{
		URLString: "",
	}
	repo, err := NewRepository(".", builder)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestDownload(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	defer testutil.CleanUpGitFixtures(t)

	fx := fixtures.Basic().One()
	builder := &mockBuilder{
		CheckoutOptions: &git.CheckoutOptions{
			Branch: plumbing.Master,
		},
		CloneOptions: &git.CloneOptions{
			URL: fx.DotGit().Root(),
		},
		URLString: fx.DotGit().Root(),
	}

	fs := memfs.New()
	s := memory.NewStorage()

	repo, err := NewRepository(".", builder)
	require.NoError(t, err)
	repo.Driver.SetFilesystem(fs)
	repo.Driver.SetStorer(s)

	err = repo.Download(false)
	assert.NoError(t, err)

	// This should try to open the repo because it is already downloaded
	repoOpen, err := NewRepository(".", builder)
	require.NoError(t, err)
	repoOpen.Driver.SetFilesystem(fs)
	repoOpen.Driver.SetStorer(s)
	err = repoOpen.Download(false)
	assert.NoError(t, err)
	ref, err := repo.Driver.Head()
	require.NoError(t, err)
	assert.NotNil(t, ref.String())
}

func TestUpdate(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	defer testutil.CleanUpGitFixtures(t)

	fx := fixtures.Basic().One()

	checkout := &git.CheckoutOptions{
		Branch: plumbing.Master,
	}
	builder := &mockBuilder{
		CheckoutOptions: checkout,
		CloneOptions: &git.CloneOptions{
			URL: fx.DotGit().Root(),
		},
		FetchOptions: &git.FetchOptions{Auth: nil},
		AuthMethod:   nil,
		URLString:    fx.DotGit().Root(),
	}

	repo, err := NewRepository(".", builder)
	require.NoError(t, err)
	driver := &GitDriver{
		Filesystem: memfs.New(),
		Storer:     memory.NewStorage(),
	}
	// Set inmemory fs instead of real one
	repo.Driver = driver
	require.NoError(t, err)

	// Clone repo into memory fs
	err = repo.Clone()
	require.NoError(t, err)
	// Get hash of the HEAD
	ref, err := repo.Driver.Head()
	require.NoError(t, err)
	headHash := ref.Hash()

	// calculate previous commit hash
	prevCommitHash, err := repo.Driver.ResolveRevision("HEAD~1")
	require.NoError(t, err)
	require.NotEqual(t, prevCommitHash.String(), headHash.String())
	builder.CheckoutOptions = &git.CheckoutOptions{Hash: *prevCommitHash}
	// Checkout previous commit
	err = repo.Checkout(true)
	require.NoError(t, err)

	// Set checkout back to master
	builder.CheckoutOptions = checkout
	err = repo.Checkout(true)
	assert.NoError(t, err)
	// update repository
	require.NoError(t, repo.Update(true))

	currentHash, err := repo.Driver.Head()
	assert.NoError(t, err)
	// Make sure that current has is same as master hash
	assert.Equal(t, headHash.String(), currentHash.Hash().String())

	repo.Driver.Close()
	updateError := repo.Update(true)
	assert.Error(t, updateError)
}

func TestOpen(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	defer testutil.CleanUpGitFixtures(t)

	fx := fixtures.Basic().One()
	url := fx.DotGit().Root()
	checkout := &git.CheckoutOptions{Branch: plumbing.Master}
	builder := &mockBuilder{
		CheckoutOptions: checkout,
		URLString:       url,
		CloneOptions:    &git.CloneOptions{Auth: nil, URL: url},
	}

	repo, err := NewRepository(".", builder)
	require.NoError(t, err)
	repo.Driver = &GitDriver{
		Filesystem: memfs.New(),
		Storer:     memory.NewStorage(),
	}

	err = repo.Clone()
	assert.NotNil(t, repo.Driver)
	require.NoError(t, err)

	// This should open the repo
	repoOpen, err := NewRepository(".", builder)
	require.NoError(t, err)

	storer := memory.NewStorage()
	err = storer.SetReference(plumbing.NewReferenceFromStrings("HEAD", ""))
	require.NoError(t, err)
	repoOpen.Driver = &GitDriver{
		Filesystem: memfs.New(),
		Storer:     storer,
	}

	require.NoError(t, repoOpen.Open())
	ref, err := repo.Driver.Head()
	assert.NoError(t, err)
	assert.NotNil(t, ref.String())
}

func TestCheckout(t *testing.T) {
	err := fixtures.Init()
	require.NoError(t, err)
	defer testutil.CleanUpGitFixtures(t)

	fx := fixtures.Basic().One()
	url := fx.DotGit().Root()
	checkout := &git.CheckoutOptions{Branch: plumbing.Master}
	builder := &mockBuilder{
		CheckoutOptions: checkout,
		URLString:       url,
		CloneOptions:    &git.CloneOptions{Auth: nil, URL: url},
	}

	repo, err := NewRepository(".", builder)
	require.NoError(t, err)
	err = repo.Checkout(true)
	assert.Error(t, err)
}

func TestURLtoName(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "https://github.com/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "https://github.com/kubernetes/kube.somepath.ctl.git",
			expectedOutput: "kube.somepath.ctl",
		},
		{
			input:          "https://github.com/kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "git@github.com:kubernetes/kubectl",
			expectedOutput: "kubectl",
		},
		{
			input:          "github.com:kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "/kubernetes/kubectl.git",
			expectedOutput: "kubectl",
		},
		{
			input:          "/kubernetes/kubectl.git/",
			expectedOutput: "",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedOutput, nameFromURL(test.input))
	}
}
