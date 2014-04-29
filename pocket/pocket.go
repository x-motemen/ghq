package pocket

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/motemen/ghq/utils"
)

var consumerKey = "27000-330870baad7ffbb5ab2fa6b2"

var apiOrigin = "https://getpocket.com"

func ObtainRequestToken(redirectURL string) (string, error) {
	utils.Log("pocket", "Obtaining request token")

	resp, err := http.PostForm(
		apiOrigin+"/v3/oauth/request",
		url.Values{
			"consumer_key": {consumerKey},
			"redirect_uri": {redirectURL},
		},
	)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Got response %d", resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	values, err := url.ParseQuery(string(buf))
	if err != nil {
		return "", err
	}

	return values.Get("code"), nil
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

	resp, err := http.PostForm(
		apiOrigin+"/v3/oauth/authorize",
		url.Values{
			"consumer_key": {consumerKey},
			"code":         {requestToken},
		},
	)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("Got response %d", resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	values, err := url.ParseQuery(string(buf))
	if err != nil {
		return "", "", err
	}

	accessToken := values.Get("access_token")
	username := values.Get("username")

	utils.Log("authorized", username)

	return accessToken, username, nil
}

func GenerateAuthorizationURL(token, redirectURL string) string {
	values := url.Values{"request_token": {token}, "redirect_uri": {redirectURL}}
	return fmt.Sprintf("%s/auth/authorize?%s", apiOrigin, values.Encode())
}
