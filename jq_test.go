package terraform_extension

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJQ(t *testing.T) {
	resp, err := runJQ("[.[].VpcId]", []interface{}{}, false)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(resp.([]interface{})), 0)
}
