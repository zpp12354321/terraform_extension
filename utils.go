package terraform_extension

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
)

func hashStringArray(arr []interface{}) string {
	var buf bytes.Buffer
	for _, s := range arr {
		buf.WriteString(fmt.Sprintf("%s-", s.(string)))
	}
	return fmt.Sprintf("%d", hashcode.String(buf.String()))
}

func isNil(c interface{}) bool {
	if c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()) {
		return true
	}
	return false
}
