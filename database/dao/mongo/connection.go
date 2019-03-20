package mongo

import (
	"net/url"
)

const connectionScheme = "mongodb"

type Connection struct {
	Host     string
	Database string
	User     string
	Password string
}

func (c Connection) String() (s string) {
	if c.Database == "" {
		return ""
	}

	vv := url.Values{}

	var userInfo *url.Userinfo

	if c.User != "" {
		if c.Password == "" {
			userInfo = url.User(c.User)
		} else {
			userInfo = url.UserPassword(c.User, c.Password)
		}
	}

	u := url.URL{
		Scheme:   connectionScheme,
		Path:     c.Database,
		Host:     c.Host,
		User:     userInfo,
		RawQuery: vv.Encode(),
	}

	return u.String()
}
