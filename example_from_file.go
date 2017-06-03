package main

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"

	"github.com/marcsantiago/OX3-Go-API-Client/openx"
	log "github.com/sirupsen/logrus"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Could not get home path:\n%v", err)
	}
	path := path.Join(usr.HomeDir, "openx_config.json")
	// Create the client
	client, err := openx.NewClientFromFile(path, false)
	if err != nil {
		log.Fatalf("Client could not be created:\n %v", err)
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
}
