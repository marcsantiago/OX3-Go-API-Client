# openx
`import "github.com/marcsantiago/OX3-Go-API-Client/openx"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [func CreateConfigFileTemplate(fileCreationPath string) string](#CreateConfigFileTemplate)
* [type Client](#Client)
  * [func NewClient(creds Credentials, debug bool) (*Client, error)](#NewClient)
  * [func NewClientFromFile(filePath string, debug bool) (*Client, error)](#NewClientFromFile)
  * [func (c *Client) Delete(url string, data io.Reader) (*http.Response, error)](#Client.Delete)
  * [func (c *Client) Get(url string, urlParms map[string]interface{}) (*http.Response, error)](#Client.Get)
  * [func (c *Client) LogOff() (res *http.Response, err error)](#Client.LogOff)
  * [func (c *Client) Options() (*http.Response, error)](#Client.Options)
  * [func (c *Client) Post(url string, data io.Reader) (*http.Response, error)](#Client.Post)
  * [func (c *Client) PostForm(url string, data url.Values) (*http.Response, error)](#Client.PostForm)
  * [func (c *Client) Put(url string, data io.Reader) (*http.Response, error)](#Client.Put)
* [type Credentials](#Credentials)


#### <a name="pkg-files">Package files</a>
[openx.go](/src/github.com/marcsantiago/OX3-Go-API-Client/openx/openx.go) 



## <a name="pkg-variables">Variables</a>
``` go
var (
    // ErrParameter definitions
    ErrParameter = errors.New("The value entered must be of type string, int, float64, or bool")
)
```


## <a name="CreateConfigFileTemplate">func</a> [CreateConfigFileTemplate](/src/target/openx.go?s=9184:9245#L362)
``` go
func CreateConfigFileTemplate(fileCreationPath string) string
```
CreateConfigFileTemplate creates a templated json file used in NewClientFromFile.
Otherwise the file format for NewClientFromFile is


	  {
		"domain": "enter domain",
		"realm": "enter realm",
		"consumer_key": "enter key",
		"consumer_secrect": "enter secrect key",
		"email": "enter email",
		"password": "enter password"
	  }

the fileCreationPath is returned incase a path is needed




## <a name="Client">type</a> [Client](/src/target/openx.go?s=1967:2265#L78)
``` go
type Client struct {
    // contains filtered or unexported fields
}
```
Client holds all the auth data and wraps calls around Go's *http.Client
Concurrency is left up to the user







### <a name="NewClient">func</a> [NewClient](/src/target/openx.go?s=2324:2386#L93)
``` go
func NewClient(creds Credentials, debug bool) (*Client, error)
```
NewClient creates the basic Openx3 *Client via oauth1


### <a name="NewClientFromFile">func</a> [NewClientFromFile](/src/target/openx.go?s=4266:4334#L167)
``` go
func NewClientFromFile(filePath string, debug bool) (*Client, error)
```
NewClientFromFile parses a JSON file to grab your Openx creds





### <a name="Client.Delete">func</a> (\*Client) [Delete](/src/target/openx.go?s=5523:5598#L219)
``` go
func (c *Client) Delete(url string, data io.Reader) (*http.Response, error)
```
Delete creates a delete request




### <a name="Client.Get">func</a> (\*Client) [Get](/src/target/openx.go?s=4866:4955#L188)
``` go
func (c *Client) Get(url string, urlParms map[string]interface{}) (*http.Response, error)
```
Get is simailiar to the normal Go *http.client.Get,
except string parameters can be passed in the url or the as a map[string]interface{}




### <a name="Client.LogOff">func</a> (\*Client) [LogOff](/src/target/openx.go?s=6956:7013#L274)
``` go
func (c *Client) LogOff() (res *http.Response, err error)
```
LogOff sets the created session to an empty http.client




### <a name="Client.Options">func</a> (\*Client) [Options](/src/target/openx.go?s=5877:5927#L232)
``` go
func (c *Client) Options() (*http.Response, error)
```
Options is a wrapper for a GET request that has the /options endpoint already passed in




### <a name="Client.Post">func</a> (\*Client) [Post](/src/target/openx.go?s=6442:6515#L256)
``` go
func (c *Client) Post(url string, data io.Reader) (*http.Response, error)
```
Post is a wrapper for the basic Go *http.client.Post, however content type is automatically set to application/json




### <a name="Client.PostForm">func</a> (\*Client) [PostForm](/src/target/openx.go?s=6707:6785#L265)
``` go
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error)
```
PostForm is a wrapper for the basic Go *http.client.PostForm




### <a name="Client.Put">func</a> (\*Client) [Put](/src/target/openx.go?s=6064:6136#L241)
``` go
func (c *Client) Put(url string, data io.Reader) (*http.Response, error)
```
Put creates a put request




## <a name="Credentials">type</a> [Credentials](/src/target/openx.go?s=950:1233#L45)
``` go
type Credentials struct {
    Domain          string `json:"domain"`
    Realm           string `json:"realm"`
    ConsumerKey     string `json:"consumer_key"`
    ConsumerSecrect string `json:"consumer_secrect"`
    Email           string `json:"email"`
    Password        string `json:"password"`
}
```
Credentials are to filled in order to auth into openx














- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
