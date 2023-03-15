package examples

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	tfext "github.com/zpp12354321/terraform_extension"
)

type VpcResource struct {
}

func (s *VpcResource) ReadVpcs(meta interface{}, request map[string]interface{}) (data []interface{}, err error) {
	// read vpcs
	return nil, nil
}

func (s *VpcResource) ReadVpc(meta interface{}, id string) (data map[string]interface{}, err error) {
	var (
		results []interface{}
		ok      bool
	)
	req := map[string]interface{}{
		"VpcIds.1": id,
	}
	results, err = s.ReadVpcs(meta, req)
	if err != nil {
		return data, err
	}
	for _, v := range results {
		if data, ok = v.(map[string]interface{}); !ok {
			return data, errors.New("Value is not map ")
		}
	}
	if len(data) == 0 {
		return data, fmt.Errorf("Vpc %s not exist ", id)
	}
	return data, err
}

func (s *VpcResource) Refresh(meta interface{}, id string) (result interface{}, state string, err error) {
	element, err := s.ReadVpc(meta, id)
	if err != nil {
		return nil, "", err
	}
	status, err := tfext.Select(".Status", element)
	if err != nil {
		return nil, "", err
	}
	return element, status.(string), nil
}

func (s *VpcResource) Datasource(d *schema.ResourceData) []tfext.Action {
	return []tfext.Action{
		{
			Name: "DescribeVpcs",
			RequestAdapter: &tfext.Adapter{
				Transformer: tfext.Transformer{
					StrCaseStyle: tfext.Camel,
					CustomTransformer: map[string]tfext.FieldTransformer{
						"ids": {
							TargetField: "VpcIds",
						},
						"tags": {
							TargetField: "TagFilters",
						},
					},
				},
				Converters: []tfext.Converter{
					tfext.NewValueConverter(".TagFilters", ".TagFilters", func(i interface{}) interface{} {
						tags := i.([]interface{})

						res := make([]interface{}, 0)
						for _, tag := range tags {
							ele := tag.(map[string]interface{})
							res = append(res, map[string]interface{}{
								"Key": ele["Key"],
								"Values": []interface{}{
									ele["Value"],
								},
							})
						}
						return res
					}),
				},
			},
			Handlers: []tfext.HandlerFunc{
				func(info *tfext.Info) tfext.Then {
					vpcs, err := s.ReadVpcs(info.Meta, info.GetRequestParams().(map[string]interface{}))
					if err != nil {
						info.AppendError(err)
						return tfext.ThenHalt
					}
					info.SetResponseParams(vpcs)
					return tfext.ThenContinue
				},
			},
			ResponseAdapter: &tfext.Adapter{
				Transformer: tfext.Transformer{
					StrCaseStyle: tfext.Snake,
				},
				IdSelector: "[.[].VpcId]", // extract VpcId for datasource id
				Converters: []tfext.Converter{
					tfext.NewJQConverter(`{"Vpcs": ., "TotalCount": length}`, true),
				},
			},
		},
	}
}

func (s *VpcResource) ReadResource(d *schema.ResourceData) []tfext.Action {
	return []tfext.Action{
		{
			Name: "DescribeVpcs",
			Handlers: []tfext.HandlerFunc{
				func(info *tfext.Info) tfext.Then {
					results, err := s.ReadVpc(info.Meta, d.Id())
					if err != nil {
						info.AppendError(err)
						return tfext.ThenHalt
					}
					info.SetResponseParams(results)
					return tfext.ThenContinue
				},
			},
			ResponseAdapter: &tfext.Adapter{
				Transformer: tfext.Transformer{
					StrCaseStyle: tfext.Snake,
				},
			},
		},
	}
}

func (s *VpcResource) CreateResource(d *schema.ResourceData) []tfext.Action {
	res := []tfext.Action{
		{
			Name: "CreateVpc",
			RequestAdapter: &tfext.Adapter{
				Transformer: tfext.Transformer{StrCaseStyle: tfext.Camel},
			},
			Handlers: []tfext.HandlerFunc{
				func(info *tfext.Info) tfext.Then {
					params := info.GetRequestParams().(map[string]interface{})
					fmt.Println("get request params", params)
					// call create vpc api
					resp := map[string]interface{}{
						"Result": map[string]interface{}{
							"VpcId": "1234567",
						},
					}
					info.SetResponseParams(resp)
					return tfext.ThenContinue
				},
				func(info *tfext.Info) tfext.Then {
					id, err := tfext.Select(".Result.VpcId", info.GetResponseParams())
					if err != nil {
						info.AppendError(err)
						return tfext.ThenHalt
					}
					d.SetId(id.(string))
					return tfext.ThenContinue
				},
				tfext.RefreshAction([]string{"Pending"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second,
					func(info *tfext.Info) (result interface{}, state string, err error) {
						return s.Refresh(info.Meta, d.Id())
					}),
			},
		},
	}
	return res
}

func (s *VpcResource) ModifyResource(d *schema.ResourceData) []tfext.Action {
	res := []tfext.Action{
		{
			Name: "ModifyVpcAttributes",
			RequestAdapter: &tfext.Adapter{
				Transformer: tfext.Transformer{StrCaseStyle: tfext.Camel},
				Converters: []tfext.Converter{
					// select target field
					tfext.NewJQConverter(`{Description: .Description, VpcName: .VpcName, DnsServers: .DnsServers}`, false),
					tfext.NewUnChangesConverter(d, map[string]tfext.Converter{
						"description": tfext.NewJQConverter("del(.Description)", false),
						"vpc_name":    tfext.NewJQConverter("del(.VpcName)", false),
						"dns_servers": tfext.NewJQConverter("del(.DnsServers)", false),
					}),
				},
			},
			Handlers: []tfext.HandlerFunc{
				func(info *tfext.Info) tfext.Then {
					params := info.GetRequestParams().(map[string]interface{})
					params["VpcId"] = d.Id()
					newParams, err := tfext.Flatten(params, &tfext.Options{StartIndex: 1})
					if err != nil {
						info.AppendError(err)
						return tfext.ThenHalt
					}
					fmt.Println("flatten request", newParams)
					// call ModifyVpcAttributes
					return tfext.ThenContinue
				},
				tfext.RefreshAction([]string{"Pending"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second,
					func(info *tfext.Info) (result interface{}, state string, err error) {
						return s.Refresh(info.Meta, d.Id())
					}),
			},
		},
	}
	if d.HasChange("tags") {
		req := map[string]interface{}{
			"ResourceType":  "vpc",
			"ResourceIds.1": d.Id(), // vpc id
		}
		added, removed := tfext.GetSetDiff("tags", d, TagsHash)
		if len(added.List()) > 0 {
			res = append(res, tfext.Action{
				Name: "TagResources",
				Handlers: []tfext.HandlerFunc{
					func(info *tfext.Info) tfext.Then {
						start := 1
						for _, element := range added.List() {
							tag := element.(map[string]interface{})
							req[fmt.Sprintf("Tags.%d.Key", start)] = tag["key"]
							req[fmt.Sprintf("Tags.%d.Value", start)] = tag["value"]
							start++
						}

						// call TagResources
						fmt.Println("call", req)
						return tfext.ThenContinue
					},
				},
			})
		}
		if len(removed.List()) > 0 {
			res = append(res, tfext.Action{
				Name: "UntagResources",
				Handlers: []tfext.HandlerFunc{
					func(info *tfext.Info) tfext.Then {
						start := 1
						for _, element := range removed.List() {
							tag := element.(map[string]interface{})
							req[fmt.Sprintf("Tags.%d.Key", start)] = tag["key"]
							start++
						}

						// call UnTagResources
						fmt.Println("call", req)
						return tfext.ThenContinue
					},
				},
			})
		}
	}
	return res
}

func (s *VpcResource) RemoveResource(d *schema.ResourceData) []tfext.Action {
	return []tfext.Action{
		{
			Handlers: []tfext.HandlerFunc{
				func(info *tfext.Info) tfext.Then {
					req := map[string]interface{}{
						"VpcId": d.Id(),
					}
					fmt.Println("start delete instance", req)
					return tfext.ThenContinue
				},
				tfext.RefreshAction([]string{"Pending"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second,
					func(info *tfext.Info) (result interface{}, state string, err error) {
						return s.Refresh(info.Meta, d.Id())
					}),
			},
		},
	}
}
