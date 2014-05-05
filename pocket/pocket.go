package pocket

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var consumerKey = "27000-330870baad7ffbb5ab2fa6b2"

var apiOrigin = "https://getpocket.com"

func requestAPIRaw(action string, params url.Values) (io.Reader, error) {
	req, err := http.NewRequest("POST", apiOrigin+action, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got response %d; X-Error=[%s]", resp.StatusCode, resp.Header.Get("X-Error"))
	}

	return resp.Body, nil
}

type RetrieveAPIResponse struct {
	Status int             `json:"status"`
	List   map[string]Item `json:"list"`
}

type Item struct {
	GivenURL    string `json:"given_url"`
	ResolvedURL string `json:"resolved_url"`
}

type OAuthRequestAPIResponse struct {
	Code string `json:"code"`
}

type OAuthAuthorizeAPIResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

func requestAPI(action string, params url.Values, v interface{}) error {
	r, err := requestAPIRaw(action, params)
	if err != nil {
		return err
	}

	d := json.NewDecoder(r)
	return d.Decode(v)
}

func ObtainRequestToken(redirectURL string) (*OAuthRequestAPIResponse, error) {
	res := &OAuthRequestAPIResponse{}
	err := requestAPI(
		"/v3/oauth/request",
		url.Values{
			"consumer_key": {consumerKey},
			"redirect_uri": {redirectURL},
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func StartAccessTokenReceiver() (string, <-chan struct{}, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, err
	}

	ch := make(chan struct{})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		ch <- struct{}{}
		_ = listener.Close()
	})

	go func() {
		_ = http.Serve(listener, nil)
	}()

	url := "http://" + listener.Addr().String()

	return url, ch, nil
}

func ObtainAccessToken(requestToken string) (*OAuthAuthorizeAPIResponse, error) {
	res := &OAuthAuthorizeAPIResponse{}
	err := requestAPI(
		"/v3/oauth/authorize",
		url.Values{
			"consumer_key": {consumerKey},
			"code":         {requestToken},
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GenerateAuthorizationURL(token, redirectURL string) string {
	values := url.Values{"request_token": {token}, "redirect_uri": {redirectURL}}
	return fmt.Sprintf("%s/auth/authorize?%s", apiOrigin, values.Encode())
}

func RetrieveGitHubEntries(accessToken string) (*RetrieveAPIResponse, error) {
	res := &RetrieveAPIResponse{}
	err := requestAPI(
		"/v3/get",
		url.Values{
			"consumer_key": {consumerKey},
			"access_token": {accessToken},
			"domain":       {"github.com"},
		},
		res,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
