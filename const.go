package terraform_extension

import "github.com/iancoleman/strcase"

type Then uint

const (
	ThenContinue Then = iota
	ThenHalt
)

type StrCaseStyle int

const (
	Camel StrCaseStyle = iota
	Snake
	LowerCamel
	Default
)

func strCase(str string, style StrCaseStyle) string {
	switch style {
	case Camel:
		return strcase.ToCamel(str)
	case Snake:
		return strcase.ToSnake(str)
	case LowerCamel:
		return strcase.ToLowerCamel(str)
	case Default:
		return str
	}
	return str
}
