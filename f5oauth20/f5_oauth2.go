// Sample client id == 5521797a84d24b27f4a4158dd56600505686379bcbc97a5c
// Sample client sec == 6e776ad034c41258f67cc7d6aeb0284ceb8acdac13e600505686379bcbc97a5c
// Look at google.ENDPOINT in main. Looks like I may just need to define an Endpoint object
// F5 Token_Endpoint == https://<host>/f5-oauth2/v1/token
// F5 Authorization Endpoint == https://<host>/f5-oauth2/v1/authorize

package f5oauth20

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/oauth2"
)

var (
//F5Endpoint is F5 specific endpoints for an OAuth virtual/host
//F5Endpoint oauth2.Endpoint
)

// F5Config describes a typical 3-legged OAuth2 flow, with both the
// client application information and the server's endpoint URLs.
type F5Config struct {
	oauth2.Config

	// Hostname is the hostname used for this session
	Hostname string

	// CAList specifies the list of client apps to pass
	CAList []string

	// State is an optional param defining weather or not to include a value for state.
	// If true, we set a random value to the state parameter
	// This should be a random value passed back to the client for verification.
	// This value protects agains CSRF attacks where a man in the middle can create a state value. If the value doesn't match then someone tampered
	State string

	// JWT is a boolean value that adds in the correct parameter for a jwt token request
	JWT bool

	// AuthURL is the full url with params to request
	AuthURL string

	// TokenURL is the full url with post params to send
	TokenURL string

	//Code holds the received token code
	Code string
}

// SetEndpoint sets the Auth and Token endpoints to the proper hostname value
func (f5c *F5Config) SetEndpoint(hostname string) (string, string) {
	hst := f5c.CleanString(hostname)
	if hst == "" {
		f5c.Hostname = "oauthas.apm.f5net.com"
	} else {
		f5c.Hostname = hst
	}

	f5c.Endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
	f5c.Endpoint.AuthURL = fmt.Sprintf("https://%s/f5-oauth2/v1/authorize", f5c.Hostname)
	f5c.Endpoint.TokenURL = fmt.Sprintf("https://%s/f5-oauth2/v1/token", f5c.Hostname)

	return f5c.Endpoint.AuthURL, f5c.Endpoint.TokenURL
}

//AuthCodeURL overrides oauth2.AuthCodeURL and adds jwt params if needed as well as other f5 specific params like ca_list
func (f5c *F5Config) AuthCodeURL() {
	var sb strings.Builder // Used to build the query
	// params to encode the Redirect URI.
	// did not do all params like this because it sorts and Bigip may want response_type first. Need to test this
	params := url.Values{}

	sb.WriteString(fmt.Sprintf("%s?", f5c.Endpoint.AuthURL))
	sb.WriteString("response_type=code&")
	sb.WriteString(fmt.Sprintf("state=%s&", f5c.State))
	sb.WriteString(fmt.Sprintf("client_id=%s&", f5c.ClientID))
	sb.WriteString(fmt.Sprintf("scope=%s&", strings.Join(f5c.Scopes, " ")))

	params.Add("redirect_uri", f5c.RedirectURL)
	sb.WriteString(fmt.Sprintf("%s&", params.Encode()))

	if f5c.JWT != false {
		sb.WriteString("token_content_type=jwt")
	}

	f5c.AuthURL = sb.String()
}

/* // TokenCodeURL currently is only called when wanting a JWT Token
func (f5c *F5Config) TokenCodeURL(c string) {
	var sb strings.Builder // Used to build the query
	// params to encode the Redirect URI.
	// did not do all params like this because it sorts and Bigip may want response_type first. Need to test this
	params := url.Values{}

	if c == "" {
		log.Fatal("Missing Code")
	}

	f5c.Code = c

	sb.WriteString(fmt.Sprintf("%s?", f5c.Endpoint.TokenURL))
	sb.WriteString("grant_type=authorization_code&")
	sb.WriteString(fmt.Sprintf("code=%s&", f5c.Code))

	params.Add("redirect_uri", f5c.RedirectURL) // Not sure I need redirect url for this
	sb.WriteString(fmt.Sprintf("%s&", params.Encode()))

	sb.WriteString(fmt.Sprintf("client_id=%s&", f5c.ClientID))
	sb.WriteString(fmt.Sprintf("client_id=%s&", f5c.ClientSecret))

	if f5c.JWT != false {
		sb.WriteString("token_content_type=jwt")
	}

	f5c.TokenURL = sb.String()
} */

// CleanString trims additional sapce from a string and returns the cleaned string
func (f5c *F5Config) CleanString(str string) string {
	if str == "" {
		return str
	}
	return strings.TrimSpace(str)
}

// CheckState will check if state should be included or not.
// If checked, generate a value to use for the state param so that it's not easily guessed
func (f5c *F5Config) CheckState(st string) string {
	var state = f5c.CleanString(st)
	if state == "" {
		return state
	}
	return genRandomString(10)
}

// Taken from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func genRandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
