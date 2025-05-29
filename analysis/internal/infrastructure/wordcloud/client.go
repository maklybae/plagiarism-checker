package wordcloud

import (
	"context"
	"fmt"
	"io"
	"time"

	"resty.dev/v3"
)

const baseURL = "https://quickchart.io"

type Client struct {
	client *resty.Client
}

func NewClient() *Client {
	return &Client{
		client: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(15*time.Second).
			SetQueryParams(map[string]string{
				"format":    "png",
				"width":     "1000",
				"height":    "1000",
				"fontScale": "15",
				"scale":     "linear",
			}).
			SetDoNotParseResponse(true).
			SetHeader("Content-Type", "application/json"),
	}
}

// DO NOT forget to close the response body after reading it.
func (c *Client) GetImage(ctx context.Context, text string) (reader io.ReadCloser, err error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("text", text).
		Get("/wordcloud")
	if err != nil {
		return nil, fmt.Errorf("failed to get wordcloud image: %w", err)
	}

	if resp.IsError() {
		if resp.RawResponse != nil {
			resp.RawResponse.Body.Close()
		}

		return nil, fmt.Errorf("wordcloud service returned error: %s", resp.Status())
	}

	// Warning: the response body is not closed here, as it is expected to be read by the caller.
	return resp.RawResponse.Body, nil
}
