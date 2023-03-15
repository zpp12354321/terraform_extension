package terraform_extension

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTrans(t *testing.T) {
	fieldsSchema := map[string]*schema.Schema{
		"availability_zone": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"ports": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"ports_empty": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"ingress": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"from": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
	}
	states := map[string]string{
		"availability_zone": "foo",
		"ports.#":           "3",
		"ports.0":           "1",
		"ports.1":           "2",
		"ports.2":           "5",
		"ingress.#":         "1",
		"ingress.0.from":    "8080",
	}
	resourceData, err := schema.InternalMap(fieldsSchema).Data(&terraform.InstanceState{
		Attributes: states,
	}, nil)
	assert.Nil(t, err)

	transformer := &Transformer{
		StrCaseStyle: Camel,
	}
	resp, err := transformer.trans(resourceData, fieldsSchema)
	assert.Nil(t, err)
	// map[AvailabilityZone:foo Ingress:[map[From:8080]] Ports:[1 2 5]]
	fmt.Println(resp)

	transformer = &Transformer{
		StrCaseStyle:   Camel,
		WithEmptyValue: true,
	}
	resp, err = transformer.trans(resourceData, fieldsSchema)
	assert.Nil(t, err)
	// map[AvailabilityZone:foo Ingress:[map[From:8080]] Ports:[1 2 5] PortsEmpty:[]]
	fmt.Println(resp)

	transformer = &Transformer{
		StrCaseStyle: Camel,
		CustomTransformer: map[string]FieldTransformer{
			"ingress": {
				TargetField: "ReplaceIngress",
				NextLevelTransformer: map[string]FieldTransformer{
					"from": {
						TargetField: "ReplaceFrom",
					},
				},
			},
		},
	}
	resp, err = transformer.trans(resourceData, fieldsSchema)
	assert.Nil(t, err)
	// map[AvailabilityZone:foo Ports:[1 2 5] ReplaceIngress:[map[ReplaceFrom:8080]]]
	fmt.Println(resp)
}

func TestSave(t *testing.T) {
	fieldsSchema := map[string]*schema.Schema{
		"availability_zone": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"ports": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"ports_empty": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"ingress": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"from": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
	}

	resourceData, err := schema.InternalMap(fieldsSchema).Data(&terraform.InstanceState{ID: "testID"}, nil)
	assert.Nil(t, err)

	data := map[string]interface{}{
		"Ports": []int{1, 2, 3, 4, 6},
		"Zone":  "fooZone",
		"Ingress": []interface{}{
			map[string]interface{}{
				"From": 9999,
			},
		},
	}

	responseTransformer := &Transformer{
		StrCaseStyle: Snake,
		CustomTransformer: map[string]FieldTransformer{
			"Zone": {
				TargetField: "availability_zone",
			},
		},
	}
	err = responseTransformer.save(resourceData, fieldsSchema, data)
	assert.Nil(t, err)

	// map[availability_zone:fooZone id:testID ingress.#:1 ingress.0.from:9999 ports.#:5 ports.0:1 ports.1:2 ports.2:3 ports.3:4 ports.4:6]
	fmt.Println(resourceData.State().Attributes)
}
