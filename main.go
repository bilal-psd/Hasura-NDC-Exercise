package main

import (
	"encoding/json"
	"log"

	"github.com/hasura/ndc-sdk-go/connector"
)

/* implementation of the Connector interface removed for brevity */

func main() {

	data, err := readConfigFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v\n", err)
		return
	}

	var config Configuration

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error unmarshaling config file: %v\n", err)
		return
	}

	if err := connector.Start[Configuration, State](&Connector{}); err != nil {
		panic(err)
	}
}

// func readConfigFile(filename string) ([]byte, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close() //Close file on exit

// 	return ioutil.ReadAll(file)
// }
