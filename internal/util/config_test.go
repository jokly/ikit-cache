package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	simpleConfig = `
URLs:
- https://golang.org
- https://www.google.com
MinTimeout: 10
MaxTimeout: 100
NumberOfRequests: 3	
`
)

func TestSimpleConfig(t *testing.T) {
	reader := strings.NewReader(simpleConfig)
	config := &Config{}
	err := config.parseConfig(reader)

	if assert.NoError(t, err) {
		assert.Len(t, config.URLs, 2)
		assert.Equal(t, 10, config.MinTimeout)
		assert.Equal(t, 100, config.MaxTimeout)
		assert.Equal(t, 3, config.NumberOfRequests)
	}
}
