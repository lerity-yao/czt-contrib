package ast

import "github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/gen/cztctl"

// ImportExpr defines import syntax for api
type ImportExpr struct {
	Import      Expr
	Value       Expr
	DocExpr     []Expr
	CommentExpr Expr
}

// VisitImportSpec implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitImportSpec(ctx *cztctl.ImportSpecContext) any {
	var list []*ImportExpr
	if ctx.ImportLit() != nil {
		lits := ctx.ImportLit().Accept(v).([]*ImportExpr)
		list = append(list, lits...)
	}
	if ctx.ImportBlock() != nil {
		blocks := ctx.ImportBlock().Accept(v).([]*ImportExpr)
		list = append(list, blocks...)
	}
	return list
}

// VisitImportLit implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitImportLit(ctx *cztctl.ImportLitContext) any {
	importToken := v.newExprWithToken(ctx.GetImportToken())
	valueExpr := ctx.ImportValue().Accept(v).(Expr)
	return []*ImportExpr{
		{
			Import:      importToken,
			Value:       valueExpr,
			DocExpr:     v.getDoc(ctx),
			CommentExpr: v.getComment(ctx),
		},
	}
}

// VisitImportBlock implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitImportBlock(ctx *cztctl.ImportBlockContext) any {
	importToken := v.newExprWithToken(ctx.GetImportToken())
	values := ctx.AllImportBlockValue()
	var list []*ImportExpr
	for _, value := range values {
		importExpr := value.Accept(v).(*ImportExpr)
		importExpr.Import = importToken
		list = append(list, importExpr)
	}
	return list
}

// VisitImportBlockValue implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitImportBlockValue(ctx *cztctl.ImportBlockValueContext) any {
	value := ctx.ImportValue().Accept(v).(Expr)
	return &ImportExpr{
		Value:       value,
		DocExpr:     v.getDoc(ctx),
		CommentExpr: v.getComment(ctx),
	}
}

// VisitImportValue implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitImportValue(ctx *cztctl.ImportValueContext) any {
	return v.newExprWithTerminalNode(ctx.STRING())
}

// Format provides a formatter for api command, now nothing to do
func (i *ImportExpr) Format() error {
	return nil
}

// Equal compares whether the element literals in two ImportExpr are equal
func (i *ImportExpr) Equal(v any) bool {
	if v == nil {
		return false
	}
	imp, ok := v.(*ImportExpr)
	if !ok {
		return false
	}
	if !EqualDoc(i, imp) {
		return false
	}
	return i.Import.Equal(imp.Import) && i.Value.Equal(imp.Value)
}

// Doc returns the document of ImportExpr
func (i *ImportExpr) Doc() []Expr {
	return i.DocExpr
}

// Comment returns the comment of ImportExpr
func (i *ImportExpr) Comment() Expr {
	return i.CommentExpr
}
