package examples

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	tfext "github.com/zpp12354321/terraform_extension"
)

func DataSourceVpcs() *schema.Resource {
	dataSchema := map[string]*schema.Schema{
		"ids": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Set: schema.HashString,
		},
		"tags": {
			Type:     schema.TypeSet,
			Optional: true,
			Set:      TagsHash,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"total_count": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"vpcs": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"vpc_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"status": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"vpc_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"cidr_block": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"description": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"subnet_ids": {
						Type:     schema.TypeSet,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Set: schema.HashString,
					},
					"project_name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"tags": {
						Type:     schema.TypeSet,
						Computed: true,
						Set:      TagsHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Type:     schema.TypeString,
									Computed: true,
								},
								"value": {
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}

	dispatcher := tfext.NewDispatcher(&VpcResource{}).WithDatasourceSchema(dataSchema)
	return &schema.Resource{
		Read:   dispatcher.Data,
		Schema: dataSchema,
	}
}
