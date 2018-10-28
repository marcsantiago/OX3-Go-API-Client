package openx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mrjones/oauth"
	"github.com/pkg/errors"
	"github.com/timehop/golog/log"
)

var (
	// ErrParameter definitions
	ErrParameter = errors.New("The value entered must be of type string, int, float64, or bool")
	// clean up entered domain just incase user passes in a domain in a way I'm not ready for
	domainReplacer = strings.NewReplacer(
		"www.", "",
		"http://", "",
		"https://", "",
		"//", "",
		"/", "",
	)

	// keeping this guy private
	consumer *oauth.Consumer
)

const (
	version          = "1.0.0"
	requestTokenURL  = "https://sso.openx.com/api/index/initiate"
	accessTokenURL   = "https://sso.openx.com/api/index/token"
	authorizationURL = "https://sso.openx.com/login/process"
	apiPath          = "/ox/4.0/"
	callBack         = "oob"
	logKey           = "Openx-Package"
)

// Credentials are to filled in order to auth into openx
type Credentials struct {
	Domain          string `json:"domain"`
	Realm           string `json:"realm"`
	ConsumerKey     string `json:"consumer_key"`
	ConsumerSecrect string `json:"consumer_secrect"`
	Email           string `json:"email"`
	Password        string `json:"password"`
}

func (c Credentials) validate() error {
	if strings.TrimSpace(c.Domain) == "" {
		return errors.New("domain cannot be emtpy")
	}
	if strings.TrimSpace(c.Realm) == "" {
		return errors.New("realm cannot be emtpy")
	}
	if strings.TrimSpace(c.ConsumerKey) == "" {
		return errors.New("consumer key cannot be emtpy")
	}
	if strings.TrimSpace(c.ConsumerSecrect) == "" {
		return errors.New("consumer secrect cannot be emtpy")
	}
	if strings.TrimSpace(c.Email) == "" {
		return errors.New("email cannot be emtpy")
	}
	if strings.TrimSpace(c.Password) == "" {
		return errors.New("password cannot be emtpy")
	}
	return nil
}

// Client holds all the auth data and wraps calls around Go's *http.Client
// Concurrency is left up to the user
type Client struct {
	domain          string
	realm           string
	scheme          string
	consumerKey     string
	consumerSecrect string
	email           string
	password        string
	apiPath         string
	session         *http.Client
}

// NewClient creates the basic Openx3 *Client via oauth1
func NewClient(creds Credentials, debug bool) (*Client, error) {
	if err := creds.validate(); err != nil {
		return nil, err
	}

	// create base client default to http
	c := &Client{
		domain:          creds.Domain,
		realm:           creds.Realm,
		consumerKey:     creds.ConsumerKey,
		consumerSecrect: creds.ConsumerSecrect,
		apiPath:         apiPath,
		email:           creds.Email,
		password:        creds.Password,
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

	accessToken, err := c.getAccessToken(debug)
	if err != nil {
		return nil, errors.Wrap(err, "Access token could not be generated")
	}

	if accessToken == nil {
		return nil, fmt.Errorf("access token is nil")
	}

	// create a cookie jar to add the access token to
	log.Trace(logKey, "Creating cookiejar")

	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Cookiejar could not be created")
	}

	c.domain = domainReplacer.Replace(c.domain)
	// format the domain
	base, err := url.Parse(fmt.Sprintf("%s://www.%s", c.scheme, c.domain))
	if err != nil {
		return nil, err
	}

	log.Trace(logKey, "setting openx3_access_token in cookie jar")

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
	log.Trace(logKey, "creating oauth1 session")

	session, err := consumer.MakeHttpClient(accessToken)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't create client")
	}
	session.Jar = cj
	c.session = session
	return c, nil
}

// NewClientFromFile parses a JSON file to grab your Openx creds
func NewClientFromFile(filePath string, debug bool) (*Client, error) {
	var creds Credentials
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't read the file: %s", filePath)
	}

	err = json.Unmarshal(contents, &creds)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't load bytes into struct")
	}

	if err := creds.validate(); err != nil {
		return nil, err
	}

	return NewClient(creds, debug)
}

// Get is simailiar to the normal Go *http.client.Get,
// except string parameters can be passed in the url or the as a map[string]interface{}
func (c *Client) Get(url string, urlParms map[string]interface{}) (*http.Response, error) {
	url, err := c.formatURL(url)
	if err != nil {
		return nil, err
	}

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
				return nil, ErrParameter
			}
			p += key + "=" + v + "&"
		}
		url += p[:len(p)-1]
	}
	return c.session.Get(url)
}

// Delete creates a delete request
func (c *Client) Delete(url string, data io.Reader) (*http.Response, error) {
	url, err := c.formatURL(url)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("DELETE", url, data)
	if err != nil {
		return nil, err
	}
	return c.session.Do(req)
}

// Options is a wrapper for a GET request that has the /options endpoint already passed in
func (c *Client) Options() (*http.Response, error) {
	url, err := c.formatURL("/options")
	if err != nil {
		return nil, err
	}
	return c.session.Get(url)
}

// Put creates a put request
func (c *Client) Put(url string, data io.Reader) (*http.Response, error) {
	url, err := c.formatURL(url)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return nil, err
	}

	return c.session.Do(req)
}

// Post is a wrapper for the basic Go *http.client.Post, however content type is automatically set to application/json
func (c *Client) Post(url string, data io.Reader) (*http.Response, error) {
	url, err := c.formatURL(url)
	if err != nil {
		return nil, err
	}
	return c.session.Post(url, "application/json", data)
}

// PostForm is a wrapper for the basic Go *http.client.PostForm
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	url, err := c.formatURL(url)
	if err != nil {
		return nil, err
	}
	return c.session.PostForm(url, data)
}

// LogOff sets the created session to an empty http.client
func (c *Client) LogOff() (res *http.Response, err error) {
	// set the session to an empty struct to clear auth information
	c.session = &http.Client{}
	return
}

func (c *Client) formatURL(endpoint string) (string, error) {
	var uri string
	rawURL, err := url.Parse(endpoint)
	if err != nil {
		return uri, err
	}
	p := strings.TrimLeft(rawURL.Path, "/")
	if rawURL.Scheme == "" {
		uri = fmt.Sprintf("%s://", c.scheme) + path.Join(c.domain, c.apiPath, p)
	}

	return uri, nil
}

func (c *Client) getAccessToken(debug bool) (*oauth.AccessToken, error) {
	requestToken, requestURL, err := consumer.GetRequestTokenAndUrl(callBack)
	if err != nil {
		return nil, err
	}

	log.Trace(logKey, "Requests Token generated")
	// auth into openx
	request := http.Client{}
	urlData := url.Values{}
	urlData.Set("email", c.email)
	urlData.Set("password", c.password)
	urlData.Set("oauth_token", requestToken.Token)

	resp, err := request.PostForm(requestURL, urlData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Trace(logKey, "Getting auth token")

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "Couldn't get authorization status returned: %d", resp.StatusCode)
	}

	io.Copy(ioutil.Discard, resp.Body)

	// parse the header, should contain oauth_verifier
	oauthVerifier := resp.Header.Get("oauth_verifier")
	if len(oauthVerifier) == 0 {
		log.Trace(logKey, "oauth_verifier is empty")
		return nil, errors.Wrap(err, "oauth_verifier is empty")
	}

	// use oauth_verifier to get access_token
	accessToken, err := consumer.AuthorizeToken(requestToken, oauthVerifier)
	log.Trace(logKey, "logging token response", "token", accessToken, "err", err)
	if err != nil {
		return nil, err
	}

	log.Trace(logKey, "Access Token generated")
	return accessToken, nil
}

// CreateConfigFileTemplate creates a templated json file used in NewClientFromFile.
// Otherwise the file format for NewClientFromFile is
/*
  {
	"domain": "enter domain",
	"realm": "enter realm",
	"consumer_key": "enter key",
	"consumer_secrect": "enter secrect key",
	"email": "enter email",
	"password": "enter password"
  }
*/
// the fileCreationPath is returned incase a path is needed
func CreateConfigFileTemplate(fileCreationPath string) string {
	configFile := `
	{
		"domain": "enter domain",
		"realm": "enter realm",
		"consumer_key": "enter key",
		"consumer_secrect": "enter secrect key",
		"email": "enter email",
		"password": "enter password"
	}
	`

	if !strings.HasSuffix(fileCreationPath, ".json") {
		fileCreationPath = path.Join(fileCreationPath, "openx_config.json")
	}

	f, err := os.Create(fileCreationPath)
	if err != nil {
		log.Error(logKey, "Couldn't create the file", "filename", fileCreationPath)
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(configFile)
	if err != nil {
		log.Error(logKey, "Couldn't write data to the file", "filename", fileCreationPath)
		panic(err)
	}

	log.Info(logKey, "The file was created", "filepath", fileCreationPath)
	return fileCreationPath
}
