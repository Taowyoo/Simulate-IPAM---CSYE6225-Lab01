package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	. "github.com/Taowyoo/Simulate-IPAM---CSYE6225-Lab01/ipamclient"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var configPath string = "config.json"

type Cfg struct {
	QueueName     string
	InitIPAddress []string
}

func readConfig() (c Cfg) {
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
	return
}

func sendIP(ipStr string, queueURL *string, client *sqs.Client) (err error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Println("Meet an invalid initial ip address:", ipStr)
		return
	}
	ipType := "ipv4"
	if len(ip) != 4 {
		ipType = "ipv6"
	}
	sMInput := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Name": {
				DataType:    aws.String("String"),
				StringValue: aws.String("IP Address"),
			},
			"Type": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(ipType),
			},
			"Timestamp": {
				DataType:    aws.String("String"),
				StringValue: aws.String(time.Now().Format(time.RFC3339)),
			},
		},
		MessageBody:    aws.String(ipStr),
		MessageGroupId: aws.String("available_ip"),
		QueueUrl:       queueURL,
	}

	_, err = SendMsg(context.TODO(), client, sMInput)
	if err != nil {
		return
	}
	fmt.Println("Sent", ip)
	return
}

func receiveIP(maxNumberOfMessages int32, queueURL *string, client *sqs.Client) (msgResult *sqs.ReceiveMessageOutput, err error) {
	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: maxNumberOfMessages,
	}
	msgResult, err = GetMessages(context.TODO(), client, gMInput)
	return
}

func deleteIP(queueURL *string, client *sqs.Client, receipt *string) (err error) {
	dMInput := &sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: receipt,
	}
	_, err = RemoveMessage(context.TODO(), client, dMInput)
	return
}

func main() {

	myCfg := readConfig()

	queue := flag.String("q", myCfg.QueueName, "The name of the queue")
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
	client := sqs.NewFromConfig(cfg)
	// Get URL of queue
	gQInput := &sqs.GetQueueUrlInput{
		QueueName: queue,
	}
	urlResult, err := GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return
	}
	queueURL := urlResult.QueueUrl

	// send initial ips

	// receive ip

	// Setup initial ips on server
	var msgResult *sqs.ReceiveMessageOutput
	msgResult, err = receiveIP(1, queueURL, client)

	fmt.Println("Succeed to delete one message.")
}
