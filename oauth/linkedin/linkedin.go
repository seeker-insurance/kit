package linkedin

import (
	"encoding/json"
	"errors"

	"github.com/parnurzeal/gorequest"
)

const (
	// DataEndpoint linkedin profile data endpoint url
	DataEndpoint = "https://api.linkedin.com/v1/people/~:(id,first-name,last-name,email-address,picture-url)"
)

var tokenTypes = map[string]string{
	"oauth":  "oauth_token",
	"oauth2": "oauth2_access_token",
}

// User linkedin user data
type User struct {
	ID         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"emailAddress"`
	PictureURL string `json:"pictureUrl"`
	Headline   string
	ErrorCode  int `json:"errorCode"`
	Message    string
	Status     int
}

// UserInfo linkedin user info by the access token
func UserInfo(tokenType string, accessToken string) (*User, error) {
	user := new(User)

	request := gorequest.New().Get(DataEndpoint)
	_, body, errs := request.
		Param(tokenTypes[tokenType], accessToken).
		Param("format", "json").
		End()
	if errs != nil {
		return user, errs[0]
	}

	if err := json.Unmarshal([]byte(body), &user); err != nil {
		return nil, err
	}

	if user.Message != "" {
		return user, errors.New(user.Message)
	}

	return user, nil
}