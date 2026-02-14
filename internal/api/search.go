package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tanq16/box/internal/client"
	"github.com/tanq16/box/internal/types"
)

// Search performs a server-side search on Box.
func Search(c *client.BoxClient, opts types.SearchOptions) (*types.SearchResults, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/search", client.APIBaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("query", opts.Query)
	q.Add("fields", "type,id,name,size,modified_at,parent")

	if opts.Type != "" {
		q.Add("type", opts.Type)
	}
	if len(opts.Extensions) > 0 {
		q.Add("file_extensions", strings.Join(opts.Extensions, ","))
	}
	if opts.FolderID != "" {
		q.Add("ancestor_folder_ids", opts.FolderID)
	}
	if opts.CreatedAfter != "" || opts.CreatedBefore != "" {
		from := opts.CreatedAfter
		to := opts.CreatedBefore
		q.Add("created_at_range", from+","+to)
	}
	if opts.UpdatedAfter != "" || opts.UpdatedBefore != "" {
		from := opts.UpdatedAfter
		to := opts.UpdatedBefore
		q.Add("updated_at_range", from+","+to)
	}
	if opts.SizeMin > 0 || opts.SizeMax > 0 {
		min := ""
		max := ""
		if opts.SizeMin > 0 {
			min = fmt.Sprintf("%d", opts.SizeMin)
		}
		if opts.SizeMax > 0 {
			max = fmt.Sprintf("%d", opts.SizeMax)
		}
		q.Add("size_range", min+","+max)
	}
	if opts.Owner != "" {
		q.Add("owner_user_ids", opts.Owner)
	}
	if opts.Sort != "" {
		q.Add("sort", opts.Sort)
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 30
	}
	q.Add("limit", fmt.Sprintf("%d", limit))

	req.URL.RawQuery = q.Encode()

	var results types.SearchResults
	_, err = c.DoJSON(req, &results)
	if err != nil {
		return nil, err
	}
	return &results, nil
}
