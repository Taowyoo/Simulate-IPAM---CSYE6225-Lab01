package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Taowyoo/Simulate-IPAM---CSYE6225-Lab01/ipamclient"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var configPath string = "config.json"
var ipPath string = "ip.json"
var myCfg cfg
var ips ipAddresses

type cfg struct {
	QueueName string
	Regin     string
	AccessKey string
	SecretKey string
}

type ipAddresses struct {
	InitIPAddress []string
}

func readJSONFile(path string) (data []byte) {
	f, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Open %s error:\n%s\n", path, err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer f.Close()
	data, err = ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("Read %s error:\n%s\n", path, err)
	}
	return
}

func initConfig() {
	cfgData := readJSONFile(configPath)
	err := json.Unmarshal(cfgData, &myCfg)
	if err != nil {
		fmt.Printf("Parse %s error:\n%s\n", configPath, err)
		return
	}
	fmt.Println("Loaded config from", configPath)
	return
}

func readIPs() {
	ipData := readJSONFile(ipPath)
	err := json.Unmarshal(ipData, &ips)
	if err != nil {
		fmt.Printf("Parse %s error:\n%s\n", ipPath, err)
		return
	}
	fmt.Println("Loaded ip addresses from", ipPath)
	return
}

func main() {

	initConfig()
	readIPs()
	queue := flag.String("q", myCfg.QueueName, "The name of the queue")
	initEnable := flag.Bool("i", false, "Whether send init ip address")
	flag.Parse()

	// Create AWS client from config
	fmt.Println("Connecting to server...")
	client := sqs.New(sqs.Options{
		Region:      myCfg.Regin,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(myCfg.AccessKey, myCfg.SecretKey, "")),
	})
	// client := sqs.NewFromConfig(cfg)
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
		ipamclient.SendInitIPs(&ips.InitIPAddress, queueURL, client)
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
			readIPs()
			ipamclient.SendInitIPs(&ips.InitIPAddress, queueURL, client)
		case "q":
			return
		}
	}

}
