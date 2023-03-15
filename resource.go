package terraform_extension

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type Resource interface {
	// Adapt For Datasource
	Datasource(*schema.ResourceData) []Action

	// Adapt For Resource
	ReadResource(*schema.ResourceData) []Action
	CreateResource(*schema.ResourceData) []Action
	ModifyResource(*schema.ResourceData) []Action
	RemoveResource(*schema.ResourceData) []Action
}
