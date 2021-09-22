package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Forks int

const (
	WithForks Forks = iota
	WithoutForks
	OnlyForks
)

type FindFileProps struct {
	Extension    string
	Forks        Forks
	User         string
	Organization string
	Repository   string

	Page            int
	IfModifiedSince time.Time
	ETag            string
}

type FindFileGithubResponse struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`

	Items []File `json:"items"`
}

type FindFileResponseMeta struct {
	Status string
	// 304 Not Modified
	// 403 Forbidden
	// 422 Un processable Entity
	// 503 Service Unavailable
	StatusCode   int
	LastModified time.Time
	ETag         string

	TotalCount        int
	IncompleteResults bool
}

func FindFile(ctx context.Context, props FindFileProps) (*FindFileResponseMeta, []File, error) {
	req, err := FindFileRequest(ctx, props)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build FindFileRequest: %w", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make find request: %w", err)
	}
	return FindFileHandleResponse(ctx, res)
}

func FindFileRequest(ctx context.Context, props FindFileProps) (*http.Request, error) {
	var query []string
	if props.Extension != "" {
		query = append(query, "extension:"+props.Extension)
	}

	if props.User != "" {
		query = append(query, "user:"+props.User)
	}

	if props.Organization != "" {
		query = append(query, "org:"+props.Organization)
	}
	if props.Repository != "" {
		query = append(query, "repo:"+props.Repository)
	}

	switch props.Forks {
	case WithForks:
		query = append(query, "fork:true")
	case OnlyForks:
		query = append(query, "fork:only")
		//case WithoutForks: // default
	}

	values := url.Values{}
	values.Add("q", strings.Join(query, " "))
	values.Add("sort", "indexed")
	values.Add("per_page", "100")
	if props.Page != 0 {
		values.Add("page", strconv.Itoa(props.Page))
	}

	href := &url.URL{Scheme: "https", Host: "api.github.com", Path: "/search/code", RawQuery: values.Encode()}

	req, err := http.NewRequestWithContext(ctx, "GET", href.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext failed %w", err)
	}
	req.Header.Set("accept", "application/vnd.github.v3+json")
	if props.ETag != "" {
		req.Header.Set("ETag", props.ETag)
	}
	if !props.IfModifiedSince.IsZero() {
		req.Header.Set("If-Modified-Since", props.IfModifiedSince.Format(http.TimeFormat))
	}

	return req, nil
}

func FindFileHandleResponse(ctx context.Context, res *http.Response) (*FindFileResponseMeta, []File, error) {
	defer res.Body.Close()
	var err error

	meta := &FindFileResponseMeta{
		Status:     res.Status,
		StatusCode: res.StatusCode,
		ETag:       res.Header.Get("ETag"),
	}
	if lastModified := res.Header.Get("Last-Modified"); lastModified != "" {
		meta.LastModified, err = http.ParseTime(lastModified)
		if err != nil {
			log.Panicf("failed parse Last-Modified header from github '%s': %+v", lastModified, err)
		}
	}

	if res.StatusCode != http.StatusOK {
		return meta, nil, nil
	}

	findFileResponse := &FindFileGithubResponse{}

	if err := json.NewDecoder(res.Body).Decode(findFileResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to decode json body: %w", err)
	}
	meta.TotalCount = findFileResponse.TotalCount
	meta.IncompleteResults = findFileResponse.IncompleteResults

	return meta, findFileResponse.Items, nil
}
