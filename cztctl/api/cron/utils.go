package cron

import (
	"fmt"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
)

// GetDoc formats a doc string with comment prefix.
func GetDoc(doc string) string {
	if len(doc) == 0 {
		return ""
	}
	return "// " + strings.Trim(doc, "\"")
}

func requestGoTypeName(r spec.Route, pkg ...string) string {
	if r.RequestType == nil {
		return ""
	}
	return golangExpr(r.RequestType, pkg...)
}

func golangExpr(ty spec.Type, pkg ...string) string {
	switch v := ty.(type) {
	case spec.PrimitiveType:
		return v.RawName
	case spec.DefineStruct:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}
		if len(pkg) == 0 {
			return v.RawName
		}
		return fmt.Sprintf("%s.%s", pkg[0], strings.Title(v.RawName))
	case spec.ArrayType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}
		if len(pkg) == 0 {
			return v.RawName
		}
		return fmt.Sprintf("[]%s", golangExpr(v.Value, pkg...))
	case spec.MapType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}
		if len(pkg) == 0 {
			return v.RawName
		}
		return fmt.Sprintf("map[%s]%s", v.Key, golangExpr(v.Value, pkg...))
	case spec.PointerType:
		if len(pkg) > 1 {
			panic("package cannot be more than 1")
		}
		if len(pkg) == 0 {
			return v.RawName
		}
		return fmt.Sprintf("*%s", golangExpr(v.Type, pkg...))
	case spec.InterfaceType:
		return v.RawName
	}
	return ""
}
