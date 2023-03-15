package examples

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	tfext "github.com/zpp12354321/terraform_extension"
)

var TagsHash = func(v interface{}) int {
	if v == nil {
		return hashcode.String("")
	}
	m := v.(map[string]interface{})
	var (
		buf bytes.Buffer
	)
	buf.WriteString(fmt.Sprintf("%v#%v", m["key"], m["value"]))
	return hashcode.String(buf.String())
}

func ResourceVpc() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"cidr_block": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsCIDR,
		},
		"vpc_name": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"dns_servers": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
			Set: schema.HashString,
		},
		"project_name": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
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
	}

	dispatcher := tfext.NewDispatcher(&VpcResource{}).WithResourceSchema(resourceSchema)
	resource := &schema.Resource{
		Create: dispatcher.Create,
		Read:   dispatcher.Read,
		Update: dispatcher.Update,
		Delete: dispatcher.Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: resourceSchema,
	}
	return resource
}
