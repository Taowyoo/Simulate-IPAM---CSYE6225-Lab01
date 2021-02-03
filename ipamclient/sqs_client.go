package ipamclient

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const processTime int32 = 1

// SQSMessageAPI defines the interface for the GetQueueUrl function.
// We use this interface to test the function using a mocked service.
type SQSMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)

	ReceiveMessage(ctx context.Context,
		params *sqs.ReceiveMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)

	DeleteMessage(ctx context.Context,
		params *sqs.DeleteMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// GetQueueURL gets the URL of an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a GetQueueUrlOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to GetQueueUrl.
func GetQueueURL(c context.Context, api *sqs.Client, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}

// SendMsg sends a message to an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a SendMessageOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to SendMessage.
func SendMsg(c context.Context, api SQSMessageAPI, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}

// GetMessages gets the most recent message from an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a ReceiveMessageOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to ReceiveMessage.
func GetMessages(c context.Context, api SQSMessageAPI, input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	return api.ReceiveMessage(c, input)
}

// RemoveMessage deletes a message from an Amazon SQS queue.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a DeleteMessageOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to DeleteMessage.
func RemoveMessage(c context.Context, api SQSMessageAPI, input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return api.DeleteMessage(c, input)
}

func SendIP(ipStr string, queueURL *string, client *sqs.Client) (err error) {
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
				DataType:    aws.String("String"),
				StringValue: aws.String(ipType),
			},
			"Timestamp": {
				DataType:    aws.String("String"),
				StringValue: aws.String(time.Now().Format(time.RFC3339)),
			},
		},
		MessageBody:    aws.String(ip.String()),
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

func ReceiveIP(maxNumberOfMessages int32, queueURL *string, client *sqs.Client) (msgResult *sqs.ReceiveMessageOutput, err error) {
	attID := strconv.Itoa(time.Now().Nanosecond())
	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:                queueURL,
		MaxNumberOfMessages:     maxNumberOfMessages,
		VisibilityTimeout:       processTime,
		ReceiveRequestAttemptId: &attID,
	}
	msgResult, err = GetMessages(context.TODO(), client, gMInput)
	return
}

func DeleteIP(queueURL *string, client *sqs.Client, receipt *string) (err error) {
	dMInput := &sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: receipt,
	}
	_, err = RemoveMessage(context.TODO(), client, dMInput)
	return
}

func SendInitIPs(ips *[]string, queueURL *string, client *sqs.Client) {
	for _, ip := range *ips {
		err := SendIP(ip, queueURL, client)
		if err != nil {
			fmt.Println("Got an error sending message:")
			fmt.Println(err)
		}
	}
}
