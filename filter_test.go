package github

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFindFile(t *testing.T) {
	meta, files, err := FindFile(context.Background(), FindFileProps{Extension: "pegilang", Organization: "pegilang"})
	require.Nil(t, err)
	assert.NotNil(t, meta)
	assert.Equal(t, 200, meta.StatusCode)
	assert.Less(t, 3, meta.TotalCount)
	assert.Less(t, 3, len(files))
	assert.NotEmpty(t, files[0].GitUrl)
}

func TestFindFileHandleResponse(t *testing.T) {

}

func TestFindFileRequest(t *testing.T) {
	req, err := FindFileRequest(context.Background(), FindFileProps{Extension: "pegilang", Organization: "pegilang"})
	require.Nil(t, err)
	assert.Contains(t, req.URL.String(), "https://api.github.com/search/code?")
	q := strings.Split(req.URL.Query().Get("q"), " ")
	assert.Contains(t, q, "extension:pegilang")
	assert.Contains(t, q, "org:pegilang")
}
