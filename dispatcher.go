package terraform_extension

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Dispatcher struct {
	resource Resource

	ResourceSchema   map[string]*schema.Schema
	DatasourceSchema map[string]*schema.Schema
}

func NewDispatcher(resource Resource) *Dispatcher {
	return &Dispatcher{
		resource: resource,
	}
}

func (d *Dispatcher) WithResourceSchema(r map[string]*schema.Schema) *Dispatcher {
	d.ResourceSchema = r
	return d
}

func (d *Dispatcher) WithDatasourceSchema(data map[string]*schema.Schema) *Dispatcher {
	d.DatasourceSchema = data
	return d
}

func (d *Dispatcher) Data(resourceData *schema.ResourceData, meta interface{}) error {
	actions := d.resource.Datasource(resourceData)
	for _, action := range actions {
		if err := d.handleAction(resourceData, meta, action, d.DatasourceSchema, false); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Read(resourceData *schema.ResourceData, meta interface{}) error {
	actions := d.resource.ReadResource(resourceData)
	for _, action := range actions {
		if err := d.handleAction(resourceData, meta, action, d.ResourceSchema, true); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) Create(resourceData *schema.ResourceData, meta interface{}) error {
	actions := d.resource.CreateResource(resourceData)
	for _, action := range actions {
		if err := d.handleAction(resourceData, meta, action, d.ResourceSchema, true); err != nil {
			return err
		}
	}
	return d.Read(resourceData, meta)
}

func (d *Dispatcher) Update(resourceData *schema.ResourceData, meta interface{}) error {
	actions := d.resource.ModifyResource(resourceData)
	for _, action := range actions {
		if err := d.handleAction(resourceData, meta, action, d.ResourceSchema, true); err != nil {
			return err
		}
	}
	return d.Read(resourceData, meta)
}

func (d *Dispatcher) Delete(resourceData *schema.ResourceData, meta interface{}) error {
	actions := d.resource.RemoveResource(resourceData)
	for _, action := range actions {
		if err := d.handleAction(resourceData, meta, action, d.ResourceSchema, true); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) handleAction(resourceData *schema.ResourceData, meta interface{}, action Action, schema map[string]*schema.Schema, isResource bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error in action: %s, reason: %v", action.Name, r)
		}
	}()
	if action.responseParams == nil {
		action.responseParams = map[string]interface{}{}
	}
	if action.requestParams == nil {
		action.requestParams = map[string]interface{}{}
	}

	if action.RequestAdapter != nil {
		if action.RequestAdapter.Transformer.CustomTransformer == nil {
			action.RequestAdapter.Transformer.CustomTransformer = map[string]FieldTransformer{}
		}
		params, err := action.RequestAdapter.adaptRequest(resourceData, schema)
		if err != nil {
			return err
		}
		action.requestParams = params
	}

	if len(action.Handlers) > 0 {
		info := &Info{
			ResourceData:   resourceData,
			Meta:           meta,
			requestParams:  action.requestParams,
			responseParams: action.responseParams,
		}
		for _, handler := range action.Handlers {
			then := handler(info)
			if then == ThenHalt {
				if info.error != nil { // 处理失败，直接抛出错误
					return info.error
				}
				break
			}
		}
		if info.error != nil {
			return info.error
		}
		// update action params
		action.requestParams = info.requestParams
		action.responseParams = info.responseParams
	}

	if action.ResponseAdapter != nil {
		if action.ResponseAdapter.Transformer.CustomTransformer == nil {
			action.ResponseAdapter.Transformer.CustomTransformer = map[string]FieldTransformer{}
		}
		if err := action.ResponseAdapter.adaptResponse(resourceData, schema, action.responseParams, isResource); err != nil {
			return err
		}
	}
	return nil
}
