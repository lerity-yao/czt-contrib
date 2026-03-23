package ast

import (
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/gen/cztctl"
)

// KvExpr describes key-value for api
type KvExpr struct {
	Key         Expr
	Value       Expr
	DocExpr     []Expr
	CommentExpr Expr
}

// VisitKvLit implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitKvLit(ctx *cztctl.KvLitContext) any {
	var kvExpr KvExpr
	kvExpr.Key = v.newExprWithToken(ctx.GetKey())
	commentExpr := v.getComment(ctx)
	if ctx.GetValue() != nil {
		valueCtx := ctx.GetValue()
		valueText := valueCtx.GetText()
		start := valueCtx.GetStart()
		valueExpr := v.newExprWithText(valueText, start.GetLine(), start.GetColumn(), start.GetStart(), valueCtx.GetStop().GetStop())
		if strings.Contains(valueText, "//") {
			if commentExpr == nil {
				commentExpr = v.newExprWithText("", start.GetLine(), start.GetColumn(), start.GetStart(), valueCtx.GetStop().GetStop())
			}
			index := strings.Index(valueText, "//")
			commentExpr.SetText(valueText[index:])
			valueExpr.SetText(strings.TrimSpace(valueText[:index]))
		} else if strings.Contains(valueText, "/*") {
			if commentExpr == nil {
				commentExpr = v.newExprWithText("", start.GetLine(), start.GetColumn(), start.GetStart(), valueCtx.GetStop().GetStop())
			}
			index := strings.Index(valueText, "/*")
			commentExpr.SetText(valueText[index:])
			valueExpr.SetText(strings.TrimSpace(valueText[:index]))
		}
		kvExpr.Value = valueExpr
	}

	kvExpr.DocExpr = v.getDoc(ctx)
	kvExpr.CommentExpr = commentExpr
	return &kvExpr
}

// Format provides a formatter for api command, now nothing to do
func (k *KvExpr) Format() error {
	return nil
}

// Equal compares whether the element literals in two KvExpr are equal
func (k *KvExpr) Equal(v any) bool {
	if v == nil {
		return false
	}
	kv, ok := v.(*KvExpr)
	if !ok {
		return false
	}
	if !EqualDoc(k, kv) {
		return false
	}
	return k.Key.Equal(kv.Key) && k.Value.Equal(kv.Value)
}

// Doc returns the document of KvExpr
func (k *KvExpr) Doc() []Expr {
	return k.DocExpr
}

// Comment returns the comment of KvExpr
func (k *KvExpr) Comment() Expr {
	return k.CommentExpr
}
