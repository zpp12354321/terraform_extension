package terraform_extension

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Adapter struct {
	resourceData *schema.ResourceData
	fieldsSchema map[string]*schema.Schema

	// IdSelector is used for Datasource
	IdSelector string
	// Transformer trans HCL to Unstructured Map
	Transformer Transformer
	// Converters filter and process the Unstructured Map,
	// so that the final result completely matches the final API request data
	Converters []Converter
}

func (a *Adapter) adaptRequest(resourceData *schema.ResourceData, fieldsSchema map[string]*schema.Schema) (interface{}, error) {
	a.resourceData = resourceData
	a.fieldsSchema = fieldsSchema

	unstructured, err := a.unStructure()
	if err != nil {
		return nil, fmt.Errorf("unstructure error. %s", err.Error())
	}
	converted, err := a.convert(unstructured)
	if err != nil {
		return nil, fmt.Errorf("convert error. %s", err.Error())
	}
	return converted, nil
}

func (a *Adapter) unStructure() (map[string]interface{}, error) {
	return a.Transformer.trans(a.resourceData, a.fieldsSchema)
}

func (a *Adapter) convert(params interface{}) (interface{}, error) {
	if len(a.Converters) == 0 {
		return params, nil
	}
	var err error

	for _, converter := range a.Converters {
		params, err = converter.Convert(params)
		if err != nil {
			return nil, err
		}
	}
	return params, nil
}

func (a *Adapter) adaptResponse(resourceData *schema.ResourceData, fieldsSchema map[string]*schema.Schema, params interface{}, isResource bool) error {
	// only for datasource
	id := strconv.FormatInt(time.Now().Unix(), 16)
	if !isResource {
		if len(a.IdSelector) > 0 {
			ids, err := runJQ(a.IdSelector, params, false)
			if err != nil {
				return err
			}
			if v, ok := ids.([]interface{}); ok {
				id = hashStringArray(v)
			} else {
				return fmt.Errorf("invalid id selector %s", a.IdSelector)
			}
		}
	}

	converted, err := a.convert(params)
	if err != nil {
		return err
	}
	ele, ok := converted.(map[string]interface{})
	if !ok {
		return fmt.Errorf("converted params must be map type")
	}
	if err = a.Transformer.save(resourceData, fieldsSchema, ele); err != nil {
		return err
	}
	if !isResource {
		resourceData.SetId(id)
	}
	return nil
}
