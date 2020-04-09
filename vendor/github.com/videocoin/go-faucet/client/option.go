package client

type Option func(*Client)

func WithBasicAuth(user, password string) Option {
	return func(c *Client) {
		c.user = user
		c.password = password
	}
}
