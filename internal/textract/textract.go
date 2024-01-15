package textract

import (
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
	return wrapper
}

func (t TextractWrapper) ParseTextLinesFromImage() ([]string, error) {
	file, err := os.ReadFile("./test/IMG_1949.png")

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

	lines := make([]string, 0, len(resp.Blocks))

	for _, block := range resp.Blocks {
		if *block.BlockType == "LINE" {
			lines = append(lines, *block.Text)
		}
	}

	return lines, err
}
