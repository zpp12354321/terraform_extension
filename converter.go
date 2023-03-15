package terraform_extension

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Converter interface {
	Convert(input interface{}) (interface{}, error)
}

func NewJQConverter(expr string, unableFindError bool) Converter {
	return &jqConverter{
		expr:            expr,
		unableFindError: unableFindError,
	}
}

type jqConverter struct {
	expr            string
	unableFindError bool
}

func (c *jqConverter) Convert(input interface{}) (interface{}, error) {
	return runJQ(c.expr, input, c.unableFindError)
}

type ValueConvertFunc func(interface{}) interface{}

type valueConverter struct {
	selectExpr, targetExpr string
	convertFunc            ValueConvertFunc
}

func (c *valueConverter) Convert(params interface{}) (interface{}, error) {
	ele, err := runJQ(c.selectExpr, params, false)
	if err != nil {
		return nil, err
	}
	if isNil(ele) { // 如果获取不到，那么直接返回
		return params, nil
	}

	var newValue interface{}
	if c.convertFunc == nil {
		newValue = ele
	} else {
		newValue = c.convertFunc(ele)
	}

	bytes, err := json.Marshal(newValue)
	if err != nil {
		return nil, err
	}
	resExpr := c.targetExpr + fmt.Sprintf(" = (%s|fromjson)", strconv.Quote(string(bytes)))
	res, err := runJQ(resExpr, params, false)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NewValueConverter using for update data
// select data through selectExpr, update using f func, and set data to targetExpr
func NewValueConverter(selectExpr, targetExpr string, convertFunc ValueConvertFunc) Converter {
	return &valueConverter{
		selectExpr:  selectExpr,
		targetExpr:  targetExpr,
		convertFunc: convertFunc,
	}
}

func NewObj2SliceConverter(selectExpr, targetExpr string) Converter {
	return &valueConverter{
		selectExpr: selectExpr,
		targetExpr: targetExpr,
		convertFunc: func(i interface{}) interface{} {
			return []interface{}{i}
		},
	}
}

func NewSlice2ObjConverter(selectExpr, targetExpr string) Converter {
	return &valueConverter{
		selectExpr: selectExpr,
		targetExpr: targetExpr,
		convertFunc: func(i interface{}) interface{} {
			if isNil(i) {
				return i
			}
			return i.([]interface{})[0]
		},
	}
}

func NewUnChangesConverter(d *schema.ResourceData, elements map[string]Converter) Converter {
	return &unChangesConverter{
		d:        d,
		elements: elements,
	}
}

type unChangesConverter struct {
	d        *schema.ResourceData
	elements map[string]Converter
}

func (c *unChangesConverter) Convert(params interface{}) (interface{}, error) {
	for tfKey, f := range c.elements {
		if !c.d.HasChange(tfKey) {
			res, err := f.Convert(params)
			if err != nil {
				return nil, err
			}
			params = res
		}
	}
	return params, nil
}
