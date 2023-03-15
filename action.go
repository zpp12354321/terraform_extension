package terraform_extension

type Action struct {
	Name string

	RequestAdapter  *Adapter // HCL->Unstructured Data
	Handlers        []HandlerFunc
	ResponseAdapter *Adapter // Unstructured Data->HCL

	requestParams  interface{}
	responseParams interface{}
}
