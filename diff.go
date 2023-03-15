package terraform_extension

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

// GetSetDiff return add and remove elements
// github.com/volcengine/terraform-provider-volcengine/blob/master/common/common_volcengine_diff_collection.go
func GetSetDiff(key string, d *schema.ResourceData, f schema.SchemaSetFunc) (*schema.Set, *schema.Set) {
	add := new(schema.Set)
	remove := new(schema.Set)
	if d.HasChange(key) {
		ov, nv := d.GetChange(key)
		if ov == nil {
			ov = new(schema.Set)
		}
		if nv == nil {
			nv = new(schema.Set)

		}
		os := ov.(*schema.Set)
		ns := nv.(*schema.Set)

		addProbably := schema.NewSet(f, ns.Difference(os).List())
		removeProbably := schema.NewSet(f, os.Difference(ns).List())

		add = addProbably.Difference(removeProbably)
		remove = removeProbably.Difference(addProbably)
		return add, remove
	}
	return add, remove
}
