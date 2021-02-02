package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func main() {
	queue := flag.String("q", "", "The name of the queue")
	timeout := flag.Int("t", 5, "How long, in seconds, that the message is hidden from others")
	flag.Parse()

	if *queue == "" {
		fmt.Println("You must supply the name of a queue (-q QUEUE)")
		return
	}

	if *timeout < 0 {
		*timeout = 0
	}

	if *timeout > 12*60*60 {
		*timeout = 12 * 60 * 60
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)

	gQInput := &sqs.GetQueueUrlInput{
		QueueName: queue,
	}

	// Get URL of queue
	urlResult, err := GetQueueURL(context.TODO(), client, gQInput)
	if err != nil {
		fmt.Println("Got an error getting the queue URL:")
		fmt.Println(err)
		return
	}

	queueURL := urlResult.QueueUrl

	sMInput := &sqs.SendMessageInput{
		// DelaySeconds: 10,  // Not available for FIFO SQS Queue
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Name": {
				DataType:    aws.String("String"),
				StringValue: aws.String("IP Address"),
			},
			"Type": {
				DataType:    aws.String("Number"),
				StringValue: aws.String("4"),
			},
			"Timestamp": {
				DataType:    aws.String("String"),
				StringValue: aws.String(time.Now().Format(time.RFC3339)),
			},
		},
		MessageBody:    aws.String("192.168.0.100"),
		MessageGroupId: aws.String("Available IP"),
		QueueUrl:       queueURL,
	}
	fmt.Println(*sMInput.MessageBody)
	resp, err := SendMsg(context.TODO(), client, sMInput)
	if err != nil {
		fmt.Println("Got an error sending the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Sent message with ID: " + *resp.MessageId)

	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   int32(*timeout),
	}

	msgResult, err := GetMessages(context.TODO(), client, gMInput)
	if err != nil {
		fmt.Println("Got an error receiving messages:")
		fmt.Println(err)
		return
	}
	fmt.Printf("Got %v message.\n", len(msgResult.Messages))
	if len(msgResult.Messages) == 0 {
		return
	}
	PrintMsgResult(msgResult)

	dMInput := &sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: msgResult.Messages[0].ReceiptHandle,
	}

	_, err = RemoveMessage(context.TODO(), client, dMInput)
	if err != nil {
		fmt.Println("Got an error deleting the message:")
		fmt.Println(err)
		return
	}
	fmt.Println("Succeed to delete one message.")
}
