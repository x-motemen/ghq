package pocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	. "github.com/onsi/gomega"
)

func TestObtainRequestToken(t *testing.T) {
	RegisterTestingT(t)

	theCode := "4a334434-a4ac-38fa-a747-4049b4"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("code=" + theCode))
	}))
	defer ts.Close()

	apiOrigin = ts.URL

	token, err := ObtainRequestToken()

	Expect(err).To(BeNil())
	Expect(token).To(Equal(theCode))
}
