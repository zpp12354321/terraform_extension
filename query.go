package terraform_extension

type PageCall func(map[string]interface{}) ([]interface{}, error)

func WithPageNumberQuery(condition map[string]interface{}, pageSizeParam string,
	pageNumParam string, pageSize int, initPageNumber int, call PageCall) (data []interface{}, err error) {
	pageNumber := initPageNumber
	for {
		var d []interface{}
		condition[pageSizeParam] = pageSize
		condition[pageNumParam] = pageNumber
		d, err = call(condition)
		if err != nil {
			return data, err
		}
		data = append(data, d...)
		if len(d) < pageSize {
			break
		}
		pageNumber = pageNumber + 1
	}
	return data, err
}
