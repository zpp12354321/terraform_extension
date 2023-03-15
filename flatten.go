package terraform_extension

// Borrowed from https://github.com/nqd/flat/
// Support StartIndex in Options
// MIT license
import (
	"reflect"
	"strconv"
)

// Options the flatten options.
// By default: Delimiter = "."
type Options struct {
	Prefix     string
	Delimiter  string
	Safe       bool
	MaxDepth   int
	StartIndex int // support start_index, for ecs/network/as service, start from 1
}

// Flatten the map, it returns a map one level deep
// regardless of how nested the original map was.
// By default, the flatten has Delimiter = ".", and
// no limitation of MaxDepth
func Flatten(nested map[string]interface{}, opts *Options) (m map[string]interface{}, err error) {
	if opts == nil {
		opts = &Options{
			Delimiter: ".",
		}
	}
	if opts.Delimiter == "" {
		opts.Delimiter = "."
	}

	m, err = flatten(opts.Prefix, 0, nested, opts)

	return
}

func flatten(prefix string, depth int, nested interface{}, opts *Options) (flatmap map[string]interface{}, err error) {
	flatmap = make(map[string]interface{})

	switch nested := nested.(type) {
	case map[string]interface{}:
		if opts.MaxDepth != 0 && depth >= opts.MaxDepth {
			flatmap[prefix] = nested
			return
		}
		if reflect.DeepEqual(nested, map[string]interface{}{}) {
			flatmap[prefix] = nested
			return
		}
		for k, v := range nested {
			// create new key
			newKey := k
			if prefix != "" {
				newKey = prefix + opts.Delimiter + newKey
			}
			fm1, fe := flatten(newKey, depth+1, v, opts)
			if fe != nil {
				err = fe
				return
			}
			update(flatmap, fm1)
		}
	case []interface{}:
		if opts.Safe {
			flatmap[prefix] = nested
			return
		}
		if reflect.DeepEqual(nested, []interface{}{}) {
			flatmap[prefix] = nested
			return
		}
		for i, v := range nested {
			newKey := strconv.Itoa(i + opts.StartIndex) // start index
			if prefix != "" {
				newKey = prefix + opts.Delimiter + newKey
			}
			fm1, fe := flatten(newKey, depth+1, v, opts)
			if fe != nil {
				err = fe
				return
			}
			update(flatmap, fm1)
		}
	default:
		flatmap[prefix] = nested
	}
	return
}

// update is the function that update to map with from
// example:
// to = {"hi": "there"}
// from = {"foo": "bar"}
// then, to = {"hi": "there", "foo": "bar"}
func update(to map[string]interface{}, from map[string]interface{}) {
	for kt, vt := range from {
		to[kt] = vt
	}
}
