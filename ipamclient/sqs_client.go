package ipamclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSMessageAPI defines the interface for the GetQueueUrl function.
// We use this interface to test the function using a mocked service.
type SQSMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)

	SendMessageBatch(ctx context.Context,
		params *sqs.SendMessageBatchInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)

	ReceiveMessage(ctx context.Context,
		params *sqs.ReceiveMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)

	DeleteMessage(ctx context.Context,
		params *sqs.DeleteMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)

	DeleteMessageBatch(ctx context.Context,
		params *sqs.DeleteMessageBatchInput,
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
func GetQueueURL(c context.Context, api SQSMessageAPI, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
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

func SendMsgBatch(c context.Context, api SQSMessageAPI, input *sqs.SendMessageBatchInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessageBatch(c, input)
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

func RemoveMessageBatch(c context.Context, api SQSMessageAPI, input *sqs.DeleteMessageBatchInput) (*sqs.DeleteMessageOutput, error) {
	return api.DeleteMessageBatch(c, input)
}

func PrintMsgResult(msgResult *sqs.ReceiveMessageOutput) {

	for i := 0; i < len(msgResult.Messages); i++ {
		head := fmt.Sprintf("--- Message %v ---", i+1)
		println(head)
		fmt.Println("Message ID:     " + *msgResult.Messages[i].MessageId)
		fmt.Println("Message Handle: " + *msgResult.Messages[i].ReceiptHandle)
		fmt.Println("Message Body:   " + *msgResult.Messages[i].Body)
		println(strings.Repeat("-", len(head)))
	}

}