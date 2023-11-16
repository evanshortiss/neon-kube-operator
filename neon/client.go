package neon

type Client struct {
	apiKey string
}

func CreateClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}
