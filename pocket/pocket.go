package pocket

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/motemen/ghq/utils"
)

var consumerKey = "27000-330870baad7ffbb5ab2fa6b2"

var apiOrigin = "https://getpocket.com"

func requestApiReader(action string, params url.Values) (io.Reader, error) {
	resp, err := http.PostForm(apiOrigin+action, params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got response %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func requestAPIContent(action string, params url.Values) (string, error) {
	buf, err := requestApiReader(action, params)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

type RetrieveApiResponse struct {
	Status int             `json:"status"`
	List   map[string]Item `json:"list"`
}

type Item struct {
	GivenURL    string `json:"given_url"`
	ResolvedURL string `json:"resolved_url"`
}

func requestAPIJSON(action string, params url.Values, v interface{}) error {
	r, err := requestApiReader(action, params)
	if err != nil {
		return err
	}
	d := json.NewDecoder(r)
	return d.Decode(v)
}

func requestAPI(action string, params url.Values) (url.Values, error) {
	content, err := requestAPIContent(action, params)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(content)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func ObtainRequestToken(redirectURL string) (string, error) {
	utils.Log("pocket", "Obtaining request token")

	data, err := requestAPI(
		"/v3/oauth/request",
		url.Values{
			"consumer_key": {consumerKey},
			"redirect_uri": {redirectURL},
		},
	)
	if err != nil {
		return "", err
	}

	return data.Get("code"), nil
}

func StartAccessTokenReceiver() (string, <-chan bool, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", nil, err
	}

	ch := make(chan bool)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		ch <- true
		_ = listener.Close()
	})

	go func() {
		_ = http.Serve(listener, nil)
	}()

	url := "http://" + listener.Addr().String()
	utils.Log("pocket", "Waiting for Pocket authentication callback at "+url)

	return url, ch, nil
}

func ObtainAccessToken(requestToken string) (string, string, error) {
	utils.Log("pocket", "Obtaining access token")

	data, err := requestAPI(
		"/v3/oauth/authorize",
		url.Values{
			"consumer_key": {consumerKey},
			"code":         {requestToken},
		},
	)
	if err != nil {
		return "", "", err
	}

	accessToken := data.Get("access_token")
	username := data.Get("username")

	utils.Log("authorized", username)

	return accessToken, username, nil
}

func GenerateAuthorizationURL(token, redirectURL string) string {
	values := url.Values{"request_token": {token}, "redirect_uri": {redirectURL}}
	return fmt.Sprintf("%s/auth/authorize?%s", apiOrigin, values.Encode())
}

func RetrieveGitHubEntries(accessToken string) (*RetrieveApiResponse, error) {
	utils.Log("pocket", "Retrieving github.com entries")

	var res RetrieveApiResponse

	err := requestAPIJSON(
		"/v3/get",
		url.Values{
			"consumer_key": {consumerKey},
			"access_token": {accessToken},
			"domain":       {"github.com"},
		},
		&res,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
