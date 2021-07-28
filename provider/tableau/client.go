package tableau

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
)

type ClientConfig struct {
	Host       string `validate:"required,url" mapstructure:"host"`
	Username   string `validate:"required" mapstructure:"username"`
	Password   string `validate:"required" mapstructure:"password"`
	ContentURL string `validate:"required" mapstructure:"content_url"`
}

type sessionRequest struct {
	XMLName     xml.Name           `xml:"tsRequest"`
	Credentials requestCredentials `xml:"credentials"`
}

type requestCredentials struct {
	Name     string      `xml:"name,attr"`
	Password string      `xml:"password,attr"`
	Site     requestSite `xml:"site"`
}

type requestSite struct {
	ContentUrl string `xml:"contentUrl,attr"`
}

type sessionResponse struct {
	XMLName        xml.Name            `xml:"tsResponse"`
	Text           string              `xml:",chardata"`
	Xmlns          string              `xml:"xmlns,attr"`
	Xsi            string              `xml:"xsi,attr"`
	SchemaLocation string              `xml:"schemaLocation,attr"`
	Credentials    responseCredentials `xml:"credentials"`
	Error          responseError       `xml:"error"`
}

type responseError struct {
	Text    string `xml:",chardata"`
	Code    string `xml:"code,attr"`
	Summary string `xml:"summary"`
	Detail  string `xml:"detail"`
}

type responseCredentials struct {
	Text  string       `xml:",chardata"`
	Token string       `xml:"token,attr"`
	Site  responseSite `xml:"site"`
	User  responseUser `xml:"user"`
}

type responseUser struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
}

type responseSite struct {
	Text       string `xml:",chardata"`
	ID         string `xml:"id,attr"`
	ContentUrl string `xml:"contentUrl,attr"`
}

type client struct {
	baseURL *url.URL

	username     string
	password     string
	contentUrl   string
	sessionToken string
	siteID       string
	userID       string

	httpClient *http.Client
}

func newClient(config *ClientConfig) (*client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}

	c := &client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		contentUrl: config.ContentURL,
		httpClient: &http.Client{},
	}

	sessionToken, siteID, userID, err := c.getSession()
	if err != nil {
		return nil, err
	}
	c.sessionToken = sessionToken
	c.siteID = siteID
	c.userID = userID

	return c, nil
}

func (c *client) getSession() (string, string, string, error) {
	sessionRequest := &sessionRequest{
		Credentials: requestCredentials{
			Name:     c.username,
			Password: c.password,
			Site: requestSite{
				ContentUrl: c.contentUrl,
			},
		},
	}

	sessionRequestXML, err := xml.MarshalIndent(sessionRequest, " ", "  ")
	if err != nil {
		return "", "", "", err
	}

	req, err := c.newRequest(http.MethodPost, "/api/3.12/auth/signin", string(sessionRequestXML))
	if err != nil {
		return "", "", "", nil
	}

	var sessionResponse sessionResponse
	if _, err := c.do(req, &sessionResponse); err != nil {
		return "", "", "", nil
	}

	return sessionResponse.Credentials.Token, sessionResponse.Credentials.Site.ID, sessionResponse.Credentials.User.ID, nil
}

func (c *client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}
	load := body.(string)
	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer([]byte(load)))
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "text/xml")
	}
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("X-Tableau-Auth", c.sessionToken)
	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newSessionToken, _, _, err := c.getSession()
		if err != nil {
			return nil, err
		}
		c.sessionToken = newSessionToken
		req.Header.Set("X-Tableau-Auth", c.sessionToken)

		// re-do the request
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
	}

	err = xml.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
