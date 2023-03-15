package terraform_extension

import (
	"fmt"

	"github.com/itchyny/gojq"
)

// runJQ only support exporting a single selection
func runJQ(expr string, ele interface{}, unableFindError bool) (interface{}, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, err
	}
	iter := query.Run(ele)

	result, ok := iter.Next()
	if !ok {
		if unableFindError {
			return nil, fmt.Errorf("couldn't get object through expression: %s", expr)
		}
		return nil, nil
	}
	if err, ok = result.(error); ok {
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Select get element through jquery expression
func Select(expr string, input interface{}) (interface{}, error) {
	return runJQ(expr, input, true)
}
