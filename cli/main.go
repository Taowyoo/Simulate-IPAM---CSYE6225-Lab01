package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Taowyoo/Simulate-IPAM---CSYE6225-Lab01/ipamclient"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var configPath string = "config.json"

type cfg struct {
	QueueName     string
	InitIPAddress []string
}

func readConfig() (c cfg) {
	jsonFile, err := os.Open(configPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Open %s error:\n%s\n", configPath, err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Read %s error:\n%s\n", configPath, err)
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		fmt.Printf("Parse %s error:\n%s\n", configPath, err)
	}
	fmt.Println("Loaded config from", configPath)
	return
}

func main() {

	myCfg := readConfig()

	queue := flag.String("q", myCfg.QueueName, "The name of the queue")
	initEnable := flag.Bool("i", false, "Whether send init ip address")
	flag.Parse()

	if *queue == "" {
		fmt.Println("You must supply the name of a queue (-q QUEUE)")
		return
	}

	// Load AWS Config from configuration and credential files
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("AWS configuration error, " + err.Error())
	}

	// Create AWS client from config
	fmt.Println("Connecting to server...")
	client := sqs.NewFromConfig(cfg)
	fmt.Println("Server connected")

	// Get URL of queue
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: queue,
	}
	urlResult, err := ipamclient.GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return
	}
	queueURL := urlResult.QueueUrl

	// send initial ips
	if *initEnable {
		ipamclient.SendInitIPs(&myCfg.InitIPAddress, queueURL, client)
	}

	// Start the loop to receive ip
	fmt.Println("You could Enter 'a' to send some ips from config.json to server.")
	for {
		fmt.Println("Please enter 'c' to get one available ip address or 'q' to exit the program:")
		var in string
		fmt.Scanln(&in)
		switch in {
		case "c":
			recResult, err := ipamclient.ReceiveIP(1, queueURL, client)
			if err != nil {
				fmt.Println("Got an error receiving message:")
				fmt.Println(err)
			} else {
				if len(recResult.Messages) != 0 {
					fmt.Println("Got one available:\n", *recResult.Messages[0].Body)
					err = ipamclient.DeleteIP(queueURL, client, recResult.Messages[0].ReceiptHandle)
					if err != nil {
						fmt.Println("Got an error deleting message:")
						fmt.Println(err)
					}
				} else {
					fmt.Println("No available ip now!")
					fmt.Println("You could Enter 'a' to send some ips from config.json to server.")
				}
			}
		case "a":
			ipamclient.SendInitIPs(&myCfg.InitIPAddress, queueURL, client)
		case "q":
			return
		}
	}

}
