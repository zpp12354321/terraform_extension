package terraform_extension

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Action struct {
	// The Name of Action
	Name string
	// RequestAdapter trans HCL to Unstructured Map
	RequestAdapter *Adapter
	// Business Handlers
	Handlers []HandlerFunc
	// ResponseAdapter trans Unstructured Map to HCL
	ResponseAdapter *Adapter

	requestParams  interface{}
	responseParams interface{}
}

type HandlerFunc func(*Info) Then

type Info struct {
	ResourceData   *schema.ResourceData
	Meta           interface{}
	requestParams  interface{}
	responseParams interface{}
	error          error
}

func (i *Info) GetRequestParams() interface{} {
	return i.requestParams
}

func (i *Info) SetRequestParams(params interface{}) {
	i.requestParams = params
}

func (i *Info) GetResponseParams() interface{} {
	return i.responseParams
}

func (i *Info) SetResponseParams(params interface{}) {
	i.responseParams = params
}

func (i *Info) AppendError(err error) {
	i.error = err
}
