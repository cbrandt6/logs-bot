package textract

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
)

const ImageDir = "./internal/textract/images"

type TextractWrapper struct {
	session *textract.Textract
}

func NewTextractWrapper() *TextractWrapper {
	wrapper := new(TextractWrapper)
	wrapper.session = textract.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))

	clearImageDir()
	return wrapper
}

func (t TextractWrapper) ParseTextLinesFromImage(filePath string) ([]string, error) {
	file, err := os.ReadFile(filePath)

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

func clearImageDir() {
	dir, err := os.Open(ImageDir)

	if err != nil {
		fmt.Println(err)
	}

	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)

	if err != nil {
		fmt.Println(err)
	}

	// Iterate over each file and delete it
	for _, fileInfo := range fileInfos {
		filePath := filepath.Join(ImageDir, fileInfo.Name())

		// Check if it's a regular file (not a directory)
		if fileInfo.Mode().IsRegular() {
			// Delete the file
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("Error deleting file %s: %s\n", filePath, err)
			} else {
				fmt.Printf("Deleted file: %s\n", filePath)
			}
		}
	}
}
