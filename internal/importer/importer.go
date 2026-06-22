package importer

import (
	"os"

	session "github.com/autohttp/autohttp/session"
)

func ImportFixture(path string) (*session.Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseFixture(data)
}
