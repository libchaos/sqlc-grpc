package metadata

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"
)

func exprToStr(e ast.Expr) string {
	switch exp := e.(type) {
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprToStr(exp.X), exp.Sel.Name)
	case *ast.Ident:
		return exp.String()
	case *ast.StarExpr:
		return "*" + exprToStr(exp.X)
	case *ast.ArrayType:
		return "[]" + exprToStr(exp.Elt)
	default:
		panic(fmt.Sprintf("invalid type %v", exp))
	}
}

func toProtoType(typ string) string {
	if strings.HasPrefix(typ, "*") {
		return toProtoType(typ[1:])
	}
	if strings.HasPrefix(typ, "[]") {
		return "repeated " + toProtoType(typ[2:])
	}
	switch typ {
	case "sql.NullBool":
		return ".google.protobuf.BoolValue"
	case "sql.NullInt32":
		return ".google.protobuf.Int32Value"
	case "int":
		return "int64"
	case "sql.NullInt64":
		return ".google.protobuf.Int64Value"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "sql.NullFloat64":
		return ".google.protobuf.DoubleValue"
	case "sql.NullString":
		return ".google.protobuf.StringValue"
	case "sql.NullTime", "time.Time":
		return ".google.protobuf.Timestamp"
	case "uuid.UUID":
		return "string"
	default:
		return typ
	}
}

func toKebabCase(s string) string {
	var res = make([]rune, 0, len(s))
	var p = '-'
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '-')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '-', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}
	return string(res)
}

func UpperFirstCharacter(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return str
}