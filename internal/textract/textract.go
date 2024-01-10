package textract

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
)

type TextractWrapper struct {
	session *textract.Textract
}

func NewTextractWrapper() *TextractWrapper {
	wrapper := new(TextractWrapper)
	wrapper.session = textract.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))
}

func (t TextractWrapper) ParseTextFromImage() {
	file, err := os.ReadFile("sample.png")
	if err != nil {
		panic(err)
	}

	resp, err := t.session.DetectDocumentText(&textract.DetectDocumentTextInput{
		Document: &textract.Document{
			Bytes: file,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	for i := 1; i < len(resp.Blocks); i++ {
		if *resp.Blocks[i].BlockType == "WORD" {
			fmt.Println(*resp.Blocks[i].Text)
		}
	}
}
