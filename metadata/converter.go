package metadata

import (
	"fmt"
	"go/ast"
	"regexp"
	"strings"
	"unicode"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func exprToStr(e ast.Expr) (string, error) {
	switch exp := e.(type) {
	case *ast.SelectorExpr:
		x, err := exprToStr(exp.X)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%s", x, exp.Sel.Name), nil
	case *ast.Ident:
		return exp.String(), nil
	case *ast.StarExpr:
		x, err := exprToStr(exp.X)
		if err != nil {
			return "", err
		}
		return "*" + x, nil
	case *ast.ArrayType:
		elt, err := exprToStr(exp.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + elt, nil
	default:
		return "", fmt.Errorf("invalid type %T - %v", exp, exp)
	}
}

func toProtoType(typ string) string {
	if strings.HasPrefix(typ, "*") {
		return toProtoType(typ[1:])
	}
	if strings.HasPrefix(typ, "[]") && typ != "[]byte" {
		return "repeated " + toProtoType(typ[2:])
	}
	switch typ {
	case "json.RawMessage", "[]byte":
		return "bytes"
	case "sql.NullBool":
		return "google.protobuf.BoolValue"
	case "sql.NullInt32":
		return "google.protobuf.Int32Value"
	case "int":
		return "int64"
	case "int16":
		return "int32"
	case "uint16":
		return "uint32"
	case "sql.NullInt64":
		return "google.protobuf.Int64Value"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "sql.NullFloat64":
		return "google.protobuf.DoubleValue"
	case "sql.NullString":
		return "google.protobuf.StringValue"
	case "sql.NullTime", "time.Time":
		return "google.protobuf.Timestamp"
	case "uuid.UUID", "net.HardwareAddr", "net.IP":
		return "string"
	case "sql.Result":
		return "ExecResult"
	default:
		if _, elementType := originalAndElementType(typ); elementType != "" {
			return elementType
		}
		return typ
	}
}

func bindToProto(src, dst, attrName, attrType string) []string {
	res := make([]string, 0)
	switch attrType {
	case "sql.NullBool":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = wrapperspb.Bool(%s.%s.Bool) }", dst, camelCaseProto(attrName), src, attrName))
	case "sql.NullInt32":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = wrapperspb.Int32(%s.%s.Int32) }", dst, camelCaseProto(attrName), src, attrName))
	case "sql.NullInt64":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = wrapperspb.Int64(%s.%s.Int64) }", dst, camelCaseProto(attrName), src, attrName))
	case "sql.NullFloat64":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = wrapperspb.Double(%s.%s.Float64) }", dst, camelCaseProto(attrName), src, attrName))
	case "sql.NullString":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = wrapperspb.String(%s.%s.String) }", dst, camelCaseProto(attrName), src, attrName))
	case "sql.NullTime":
		res = append(res, fmt.Sprintf("if %s.%s.Valid {", src, attrName))
		res = append(res, fmt.Sprintf("%s.%s = timestamppb.New(%s.%s.Time) }", dst, camelCaseProto(attrName), src, attrName))
	case "time.Time":
		res = append(res, fmt.Sprintf("%s.%s = timestamppb.New(%s.%s)", dst, camelCaseProto(attrName), src, attrName))
	case "uuid.UUID", "net.HardwareAddr", "net.IP":
		res = append(res, fmt.Sprintf("%s.%s = %s.%s.String()", dst, camelCaseProto(attrName), src, attrName))
	case "int16":
		res = append(res, fmt.Sprintf("%s.%s = int32(%s.%s)", dst, camelCaseProto(attrName), src, attrName))
	default:
		_, elementType := originalAndElementType(attrType)
		if elementType != "" {
			res = append(res, fmt.Sprintf("%s.%s = %s(%s.%s)", dst, camelCaseProto(attrName), elementType, src, attrName))
		} else {
			res = append(res, fmt.Sprintf("%s.%s = %s.%s", dst, camelCaseProto(attrName), src, attrName))
		}
	}
	return res
}

func bindToGo(src, dst, attrName, attrType string, newVar bool) []string {
	res := make([]string, 0)
	switch attrType {
	case "sql.NullBool":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("%s = sql.NullBool{Valid: true, Bool: v.Value}", dst))
		res = append(res, "}")
	case "sql.NullInt32":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("%s = sql.NullInt32{Valid: true, Int32: v.Value}", dst))
		res = append(res, "}")
	case "sql.NullInt64":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("%s = sql.NullInt64{Valid: true, Int64: v.Value}", dst))
		res = append(res, "}")
	case "sql.NullFloat64":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("%s = sql.NullFloat64{Valid: true, Float64: v.Value}", dst))
		res = append(res, "}")
	case "sql.NullString":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("%s = sql.NullString{Valid: true, String: v.Value}", dst))
		res = append(res, "}")
	case "sql.NullTime":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("if err := v.CheckValid(); err != nil { err = fmt.Errorf(\"invalid %s: %%s%%w\", err.Error(), validation.ErrUserInput)", attrName))
		res = append(res, "return nil, err }")
		res = append(res, "if t := v.AsTime(); !t.IsZero() {")
		res = append(res, fmt.Sprintf("%s.Valid = true", dst))
		res = append(res, fmt.Sprintf("%s.Time = t } }", dst))
	case "time.Time":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v := %s.Get%s(); v != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("if err := v.CheckValid(); err != nil { err = fmt.Errorf(\"invalid %s: %%s%%w\", err.Error(), validation.ErrUserInput)", attrName))
		res = append(res, "return nil, err }")
		res = append(res, fmt.Sprintf("%s = v.AsTime()", dst))
		res = append(res, fmt.Sprintf("} else { err := fmt.Errorf(\"field %s is required%%w\", validation.ErrUserInput)", attrName))
		res = append(res, "return nil, err }")
	case "uuid.UUID":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v, err := uuid.Parse(%s.Get%s()); err != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("err = fmt.Errorf(\"invalid %s: %%s%%w\", err.Error(), validation.ErrUserInput)", attrName))
		res = append(res, fmt.Sprintf("return nil, err } else { %s = v }", dst))

	case "net.HardwareAddr":
		if newVar {
			res = append(res, fmt.Sprintf("var %s %s", dst, attrType))
		}
		res = append(res, fmt.Sprintf("if v, err = net.ParseMAC(%s.Get%s()); err != nil {", src, camelCaseProto(attrName)))
		res = append(res, fmt.Sprintf("err = fmt.Errorf(\"invalid %s: %%s%%w\", err.Error(), validation.ErrUserInput)", attrName))
		res = append(res, fmt.Sprintf("return nil, err } else { %s = v }", dst))
	case "net.IP":
		if newVar {
			res = append(res, fmt.Sprintf("%s := net.ParseIP(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		} else {
			res = append(res, fmt.Sprintf("%s = net.ParseIP(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		}
	case "int16":
		if newVar {
			res = append(res, fmt.Sprintf("%s := int16(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		} else {
			res = append(res, fmt.Sprintf("%s = int16(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		}
	case "int":
		if newVar {
			res = append(res, fmt.Sprintf("%s := int(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		} else {
			res = append(res, fmt.Sprintf("%s = int(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		}
	case "uint16":
		if newVar {
			res = append(res, fmt.Sprintf("%s := uint16(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		} else {
			res = append(res, fmt.Sprintf("%s = uint16(%s.Get%s())", dst, src, camelCaseProto(attrName)))
		}
	default:
		originalType, elementType := originalAndElementType(attrType)
		if newVar {
			if elementType != "" {
				res = append(res, fmt.Sprintf("%s := %s(%s.Get%s())", dst, originalType, src, camelCaseProto(attrName)))
			} else {
				res = append(res, fmt.Sprintf("%s := %s.Get%s()", dst, src, camelCaseProto(attrName)))
			}
		} else {
			if elementType != "" {
				res = append(res, fmt.Sprintf("%s = %s(%s.Get%s())", dst, originalType, src, camelCaseProto(attrName)))
			} else {
				res = append(res, fmt.Sprintf("%s = %s.Get%s()", dst, src, camelCaseProto(attrName)))
			}
		}
	}
	return res
}

func UpperFirstCharacter(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return str
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toKebabCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}

func camelCaseProto(str string) string {
	return generator.CamelCase(ToSnakeCase(str))
}
