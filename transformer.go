package terraform_extension

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Transformer struct {
	StrCaseStyle      StrCaseStyle
	CustomTransformer map[string]FieldTransformer
	WithEmptyValue    bool // 是否获取空值
}

type FieldTransformer struct {
	TargetField          string
	Ignore               bool
	NextLevelTransformer map[string]FieldTransformer
}

type transResult struct {
	Ignore bool
	Key    string
	Value  interface{}
}

func (t *Transformer) trans(resourceData *schema.ResourceData, fieldsSchema map[string]*schema.Schema) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	for key, fieldSchema := range fieldsSchema {
		fieldTransformer := FieldTransformer{}
		if v, exist := t.CustomTransformer[key]; exist {
			fieldTransformer = v
		}
		res := t.transField(resourceData, key, "", fieldSchema, fieldTransformer)
		if res.Ignore {
			continue
		}
		params[res.Key] = res.Value
	}
	return params, nil
}

func (t *Transformer) transField(resourceData *schema.ResourceData, key, chain string, fieldSchema *schema.Schema, fieldTransformer FieldTransformer) *transResult {
	if fieldTransformer.Ignore {
		return &transResult{
			Ignore: true,
		}
	}
	if !(fieldSchema.Optional || fieldSchema.Required) { // only for Required or Optional Field
		return &transResult{
			Ignore: true,
		}
	}
	finalKey := strCase(key, t.StrCaseStyle)
	if fieldTransformer.TargetField != "" {
		finalKey = fieldTransformer.TargetField
	}
	res := &transResult{
		Key: finalKey,
	}
	if fieldSchema.Type == schema.TypeList || fieldSchema.Type == schema.TypeSet {
		if _, ok := fieldSchema.Elem.(*schema.Resource); ok { // only for *schema.Resource
			data, ok := get(resourceData, chain+key, t.WithEmptyValue)
			if !ok {
				return &transResult{
					Ignore: true,
				}
			}

			var (
				m     *schema.Set
				isSet bool
			)
			if m, ok = data.(*schema.Set); ok {
				isSet = true
				data = m.List()
			}
			params := make([]interface{}, 0)
			for index, element := range data.([]interface{}) {
				_index := index
				if isSet {
					_index = m.F(element)
				}
				schemaChain := chain + key + "." + strconv.Itoa(_index) + "."
				ci := make(map[string]interface{})

				for iKey, iSchema := range fieldSchema.Elem.(*schema.Resource).Schema {
					iTrans := FieldTransformer{}
					if v, exist := fieldTransformer.NextLevelTransformer[iKey]; exist {
						iTrans = v
					}

					ele := t.transField(resourceData, iKey, schemaChain, iSchema, iTrans)
					if ele.Ignore {
						continue
					}
					ci[ele.Key] = ele.Value
				}

				if len(ci) == 0 { // 没有数据
					continue
				}
				params = append(params, ci)
			}
			if !t.WithEmptyValue && len(params) == 0 {
				return &transResult{
					Ignore: true,
				}
			}
			res.Value = params
			return res
		}
	}
	if v, ok := get(resourceData, chain+key, t.WithEmptyValue); ok {
		if _, ok := v.(*schema.Set); ok {
			v = v.(*schema.Set).List()
		}
		res.Value = v
		return res
	}
	return &transResult{
		Ignore: true,
	}
}

func get(d *schema.ResourceData, key string, withEmpty bool) (interface{}, bool) {
	if withEmpty {
		return d.Get(key), true
	}
	return d.GetOk(key)
}

type saveResult struct {
	Ignore bool
	Key    string
	Value  interface{}
}

func (t *Transformer) save(resourceData *schema.ResourceData, fieldsSchema map[string]*schema.Schema, params map[string]interface{}) error {
	for key, value := range params {
		fieldTransformer := FieldTransformer{}
		if v, exist := t.CustomTransformer[key]; exist {
			fieldTransformer = v
		}
		finalKey := strCase(key, t.StrCaseStyle)
		if fieldTransformer.TargetField != "" {
			finalKey = fieldTransformer.TargetField
		}
		fieldSchema, exist := fieldsSchema[finalKey]
		if !exist {
			continue
		}

		target, err := t.saveField(resourceData, key, value, fieldSchema, fieldTransformer)
		if err != nil {
			return err
		}
		if target.Ignore {
			continue
		}
		if err = resourceData.Set(target.Key, target.Value); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transformer) saveField(resourceData *schema.ResourceData, key string, value interface{}, fieldSchema *schema.Schema, fieldTransformer FieldTransformer) (*saveResult, error) {
	if fieldTransformer.Ignore {
		return &saveResult{
			Ignore: true,
		}, nil
	}
	finalKey := strCase(key, t.StrCaseStyle)
	if fieldTransformer.TargetField != "" {
		finalKey = fieldTransformer.TargetField
	}
	res := &saveResult{Key: finalKey}
	kind := reflect.ValueOf(value).Kind()
	switch kind {
	case reflect.Map: // 说明对应的是 schema.Map
		if fieldSchema.Type != schema.TypeMap {
			return nil, fmt.Errorf("invalid mapping for key: %s", key)
		}
		res.Value = value
		return res, nil
	case reflect.Slice: // 说明对应 List OR Set
		if fieldSchema.Type == schema.TypeList || fieldSchema.Type == schema.TypeSet {
			if _, ok := fieldSchema.Elem.(*schema.Schema); ok {
				res.Value = value
				return res, nil
			}
			//  (*schema.Resource)
			iSchema := fieldSchema.Elem.(*schema.Resource)
			elements := make([]interface{}, 0)
			// 对子资源进行处理
			for _, element := range value.([]interface{}) {
				if _, ok := element.(map[string]interface{}); !ok {
					return nil, fmt.Errorf("invalid mapping for key: %s", key)
				}

				ele := make(map[string]interface{})
				for iKey, iValue := range element.(map[string]interface{}) {
					iTrans := FieldTransformer{}
					if v, exist := fieldTransformer.NextLevelTransformer[iKey]; exist {
						iTrans = v
					}
					iFinalKey := strCase(iKey, t.StrCaseStyle)
					if iTrans.TargetField != "" {
						iFinalKey = iTrans.TargetField
					}
					iFieldSchema, exist := iSchema.Schema[iFinalKey]
					if !exist {
						continue
					}
					iRes, err := t.saveField(resourceData, iKey, iValue, iFieldSchema, iTrans)
					if err != nil {
						return nil, err
					}
					if iRes.Ignore {
						continue
					}
					ele[iRes.Key] = iRes.Value
				}
				elements = append(elements, ele)
			}
			res.Value = elements
			return res, nil
		} else {
			return nil, fmt.Errorf("invalid mapping for key: %s", key)
		}
	default:
		res.Value = value
		return res, nil
	}
}
