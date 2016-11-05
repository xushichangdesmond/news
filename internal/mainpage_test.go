package news

import (
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
)

func TestExtractZipFileNames(t *testing.T) {
	testCases := []struct {
		payload      string
		expectedZips []string
	}{
		{
			`<!DOCTYPE html>	
			<html>lang="en">
			<head></head>
			<body>
			<a href="notAZip.tar">qwerty</a>
			<a href="notAZipEither.tar"></a>
			<a href="someZip.zip">wahaha</a>
			<p>
			Hello World<a href="anotherZip.zip"></a>
			</p>
			</body>`,
			[]string{"someZip.zip", "anotherZip.zip"},
		},
	}
	for _, c := range testCases {
		zips, err := extractZipFileNames(strings.NewReader(c.payload))
		assert.NoError(t, err)
		assert.Equal(t, c.expectedZips, zips)
	}
}
