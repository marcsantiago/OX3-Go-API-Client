package openx

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/mrjones/oauth"
	log "github.com/sirupsen/logrus"
)

// oauth global consumer
var consumer *oauth.Consumer

const (
	version          = "0.0.1"
	requestTokenURL  = "https://sso.openx.com/api/index/initiate"
	accessTokenURL   = "https://sso.openx.com/api/index/token"
	authorizationURL = "https://sso.openx.com/login/process"
	apiPath          = "/ox/4.0/"
	callBack         = "oob"
)

// Client holds all the user information, all of it is private however and
// as a result only the defined exported methods below are used to interact with the client.
// At the moment i'm only supporting the APIPath2 and 4.0
type Client struct {
	domain          string
	realm           string
	scheme          string
	consumerKey     string
	consumerSecrect string
	email           string
	password        string
	apiPath         string
	timeOut         int
	session         *http.Client
}

// Get is simailiar to the normal Go *http.Client.Get
// except string parameters can be passed in the url or the as a map[string]interface{}
func (c *Client) Get(url string, urlParms map[string]interface{}) (res *http.Response, err error) {
	url = c.resolveURL(url)
	if urlParms != nil {
		p := "?"
		for key, value := range urlParms {
			var v string
			switch value.(type) {
			case string:
				v = value.(string)
			case int:
				v = strconv.Itoa(value.(int))
			case float64:
				v = strconv.FormatFloat(value.(float64), 'f', -1, 64)
			case bool:
				v = strconv.FormatBool(value.(bool))
			default:
				log.Fatalln("The value entered %v must be of type string, int, float64, or bool")
			}
			p += key + "=" + v + "&"
		}
		url += p[:len(p)-1]
	}
	res, err = c.session.Get(url)
	return
}

// Delete creates a delete request which the Openx3 API uses, but that is not defined by Go
func (c *Client) Delete(url string, data io.Reader) (res *http.Response, err error) {
	request, err := http.NewRequest("DELETE", c.resolveURL(url), data)
	if err != nil {
		log.Fatalf("Could not create DELETE requests: %v", err)
	}
	res, err = c.session.Do(request)
	return
}

// Options is a wrapper for a GET request that has the /options endpoint already passed in
func (c *Client) Options(url string) (res *http.Response, err error) {
	if !strings.Contains(url, "/options") {
		url = path.Join("/options", url)
	}
	res, err = c.session.Get(c.resolveURL(url))
	return
}

// Put creates a put request which the Openx3 API uses, but that is not defined by Go
func (c *Client) Put(url string, data io.Reader) (res *http.Response, err error) {
	request, err := http.NewRequest("PUT", c.resolveURL(url), data)
	if err != nil {
		log.Fatalf("Could not create PUT requests: %v", err)
	}
	res, err = c.session.Do(request)
	return
}

// Post is a wrapper for the basic Go *http.Client.Post, however content type is automatically set to application/json
// as per Openx's documentation https://docs.openx.com/Content/developers/platform_api/api_req_and_responses.html
func (c *Client) Post(url string, data io.Reader) (res *http.Response, err error) {
	res, err = c.session.Post(c.resolveURL(url), "application/json", data)
	return
}

// PostForm is a wrapper for the basic Go *http.Client.PostForm
func (c *Client) PostForm(url string, data url.Values) (res *http.Response, err error) {
	res, err = c.session.PostForm(c.resolveURL(url), data)
	return
}

// LogOff sets the created session to an empty http.Client, destorying the stored cookie
func (c *Client) LogOff() (res *http.Response, err error) {
	// set the session to an empty struct to clear auth information
	c.session = &http.Client{}
	return
}

func (c *Client) resolveURL(endpoint string) (u string) {
	rawURL, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalln("Could not parse endpoint: %v", err)
	}
	if rawURL.Scheme == "" {
		u = fmt.Sprintf("%s://", c.scheme) + path.Join(c.domain, c.apiPath, rawURL.Path)
	}
	return
}

func (c *Client) getAccessToken(debug bool) *oauth.AccessToken {
	requestToken, requestURL, err := consumer.GetRequestTokenAndUrl(callBack)
	if err != nil {
		log.Fatalf("Requests token could not be generated %v", err)
	}
	if debug {
		log.Info("Requests Token generated")
	}

	// auth into openx
	request := http.Client{}
	urlData := url.Values{}
	urlData.Set("email", c.email)
	urlData.Set("password", c.password)
	urlData.Set("oauth_token", requestToken.Token)
	resp, err := request.PostForm(requestURL, urlData)
	if err != nil {
		log.Fatalf("Could not get authorization token: %v", err)
	}
	if debug {
		log.Info("Getting auth token")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Could not get authorization status returned: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read body: %v", err)
	}
	// parse the response, should contain oauth_verifier
	raw, err := url.Parse(string(body))
	if err != nil {
		log.Fatalf("Could not parse url: %v", err)
	}
	authInfo := raw.Query()
	var oauthVerifier string
	if val, ok := authInfo["oauth_verifier"]; !ok {
		log.Fatalln("oauth_verifier key not in response", err)
	} else {
		oauthVerifier = val[0]
	}
	// use oauth_verifier to get access_token
	accessToken, err := consumer.AuthorizeToken(requestToken, oauthVerifier)
	if err != nil {
		log.Fatalf("Access token could not be generated %v", err)
	}
	if debug {
		log.Info("Access Token generated")
	}
	return accessToken
}

// NewClient creates the basic Openx3 *Client via oauth1
func NewClient(domain, realm, consumerKey, consumerSecrect, email, password string, debug bool) (*Client, error) {
	if strings.TrimSpace(domain) == "" {
		return &Client{}, fmt.Errorf("Domain cannot be emtpy")
	}
	if strings.TrimSpace(realm) == "" {
		return &Client{}, fmt.Errorf("Realm cannot be emtpy")
	}
	if strings.TrimSpace(consumerKey) == "" {
		return &Client{}, fmt.Errorf("Consumer key cannot be emtpy")
	}
	if strings.TrimSpace(consumerSecrect) == "" {
		return &Client{}, fmt.Errorf("Consumer secrect cannot be emtpy")
	}
	if strings.TrimSpace(email) == "" {
		return &Client{}, fmt.Errorf("email cannot be emtpy")
	}
	if strings.TrimSpace(password) == "" {
		return &Client{}, fmt.Errorf("password cannot be emtpy")
	}

	// create base client default to http
	c := &Client{
		domain:          domain,
		realm:           realm,
		consumerKey:     consumerKey,
		consumerSecrect: consumerSecrect,
		apiPath:         apiPath,
		email:           email,
		password:        password,
		scheme:          "http",
	}

	// create oauth consumer
	consumer = oauth.NewConsumer(c.consumerKey, c.consumerSecrect, oauth.ServiceProvider{
		RequestTokenUrl:   requestTokenURL,
		AuthorizeTokenUrl: authorizationURL,
		AccessTokenUrl:    accessTokenURL,
		HttpMethod:        "POST",
	})
	consumer.Debug(debug)

	accessToken := c.getAccessToken(debug)

	// create a cookie jar to add the access token to
	cj, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Cookiejar could not be created %v", err)
	}

	// clean up entered domain just incase user passes in a domain in a way I'm not ready for
	r := strings.NewReplacer(
		"www.", "",
		"http://", "",
		"https://", "",
		"//", "",
		"/", "",
	)

	c.domain = r.Replace(c.domain)
	// format the domain
	base, err := url.Parse(fmt.Sprintf("%s://www.%s", c.scheme, c.domain))
	if err != nil {
		log.Fatalf("Domain could not be parsed to type url %v", err)
	}

	// create auth cookie
	var cookies []*http.Cookie
	cookie := &http.Cookie{
		Name:   "openx3_access_token",
		Value:  accessToken.Token,
		Path:   "/",
		Domain: c.domain,
		Secure: false,
		// HttpOnly: false,
	}
	cookies = append(cookies, cookie)
	cj.SetCookies(base, cookies)

	// create authenticated session
	session, err := consumer.MakeHttpClient(accessToken)
	if err != nil {
		log.Fatalf("Could not create client %v", err)
	}
	session.Jar = cj
	c.session = session
	return c, nil
}
