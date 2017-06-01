package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/marcsantiago/OX3-Go-API-Client/openx"
)

// Add this data either by hard coding, from db, or environment
// doesn't need to be defined below
const (
	email    = ""
	password = ""
	key      = ""
	secret   = ""
	domain   = ""
	realm    = ""
	debug    = false
)

// Examples do not return structs at the moment but raw json responses from Openx,
// therefore I'm using generetic interfaces and reflection to access the data.
// Structs would be faster because they avoid reflection and type casting.
// The client itself does not use any concurrency, but I might implement a semphore on the client to stop users from hitting
// openx to hard by mistake...
// All example functions provided are just that, examples. You can use the client with any Openx API endpoint.
// It is recommended that you understand thier docs https://docs.openx.com/Content/developers/platform_api/about_topics_api.html

func getReports(client *openx.Client) {
	res, err := client.Get("/report/get_reportlist", nil)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	data, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(data))
	return
}

func getAdunitIds(client *openx.Client) []string {
	offset, limit := 0, 500
	hasMore := true
	set := make(map[string]bool)
	for hasMore {
		res, err := client.Get("/adunit", map[string]interface{}{"offset": offset, "limit": limit})
		if err != nil {
			log.Fatalf("Error getting ad units: %v", err)
		}
		defer res.Body.Close()
		offset += limit
		var data map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			log.Fatalf("Marshalling error: %v", err)
		}

		if val, ok := data["objects"]; ok {
			objects := val.([]interface{})
			for _, obj := range objects {
				innerObj := obj.(map[string]interface{})
				myID := innerObj["id"].(string)
				set[myID] = true
			}
		}
		if val, ok := data["has_more"]; ok {
			hasMore = val.(bool)
		}
	}
	accountIds := make([]string, 0, len(set))
	for key := range set {
		accountIds = append(accountIds, key)
	}
	return accountIds
}

func updateLineItem(client *openx.Client, uid string, putData map[string]string) (res *http.Response, err error) {
	data, _ := json.Marshal(putData)
	res, err = client.Put(fmt.Sprintf("/lineitem/%s", uid), bytes.NewReader(data))
	return
}

func updatePrice(client *openx.Client, price float64, lineItemName string) bool {
	offset, limit := 0, 500
	hasMore := true
	for hasMore {
		res, err := client.Get("/lineitem", map[string]interface{}{"offset": offset, "limit": limit})
		if err != nil {
			log.Fatalf("Error getting ad units: %v", err)
		}
		defer res.Body.Close()
		offset += limit
		var data map[string]interface{}
		err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			log.Fatalf("Marshalling error: %v", err)
		}
		if val, ok := data["objects"]; ok {
			objects := val.([]interface{})
			for _, obj := range objects {
				lineItem := obj.(map[string]interface{})
				if lineItem["name"].(string) == lineItemName {
					newPrice := strconv.FormatFloat(price, 'f', 1, 64)
					putData := map[string]string{"pricing_rate": newPrice}
					res, err := updateLineItem(client, lineItem["uid"].(string), putData)
					if err != nil {
						log.Fatalf("Error updating lineitem: %v", err)
					}
					defer res.Body.Close()
					if res.StatusCode == http.StatusOK {
						return true
					}
					return false
				}
			}
		}

		// exit inifinate loop
		if val, ok := data["has_more"]; ok {
			hasMore = val.(bool)
		}
	}
	return false
}

func main() {
	// Create the client
	client, err := openx.NewClient(domain, realm, key, secret, email, password, debug)
	if err != nil {
		log.Fatalf("Client could not be created %v", err)
	}
	// remove cookie and session information from client
	defer client.LogOff()

	// get reports example
	getReports(client)

	// list adunit ids example
	ids := getAdunitIds(client)
	for _, i := range ids {
		fmt.Println(i)
	}

	// get options example
	res, err := client.Options("/options/ad_category_options")
	if err != nil {
		log.Fatalf("Options error %v", err)
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ioutil could not readall %v", err)
	}
	fmt.Println(string(data))

	// update the price on a line item example
	worked := updatePrice(client, 1.00, "some line item name")
	fmt.Println(worked)
}
