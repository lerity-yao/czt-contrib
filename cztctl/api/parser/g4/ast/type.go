package ast

import (
	"fmt"
	"sort"

	"github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/gen/cztctl"
)

type (
	// TypeExpr describes an expression for TypeAlias and TypeStruct
	TypeExpr interface {
		Doc() []Expr
		Format() error
		Equal(v any) bool
		NameExpr() Expr
	}

	// TypeAlias describes alias ast for api syntax
	TypeAlias struct {
		Name        Expr
		Assign      Expr
		DataType    DataType
		DocExpr     []Expr
		CommentExpr Expr
	}

	// TypeStruct describes structure ast for api syntax
	TypeStruct struct {
		Name    Expr
		Struct  Expr
		LBrace  Expr
		RBrace  Expr
		DocExpr []Expr
		Fields  []*TypeField
	}

	// TypeField describes field ast for api syntax
	TypeField struct {
		IsAnonymous bool
		Name        Expr
		DataType    DataType
		Tag         Expr
		DocExpr     []Expr
		CommentExpr Expr
	}

	// DataType describes datatype for api syntax
	DataType interface {
		Expr() Expr
		Equal(dt DataType) bool
		Format() error
		IsNotNil() bool
	}

	// Literal describes the basic types of golang
	Literal struct {
		Literal Expr
	}

	// Interface describes the interface type of golang
	Interface struct {
		Literal Expr
	}

	// Map describes the map ast for api syntax
	Map struct {
		MapExpr Expr
		Map     Expr
		LBrack  Expr
		RBrack  Expr
		Key     Expr
		Value   DataType
	}

	// Array describes the slice ast for api syntax
	Array struct {
		ArrayExpr Expr
		LBrack    Expr
		RBrack    Expr
		Literal   DataType
	}

	// Pointer describes the pointer ast for api syntax
	Pointer struct {
		PointerExpr Expr
		Star        Expr
		Name        Expr
	}
)

// VisitTypeSpec implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeSpec(ctx *cztctl.TypeSpecContext) any {
	if ctx.TypeLit() != nil {
		return []TypeExpr{ctx.TypeLit().Accept(v).(TypeExpr)}
	}
	return ctx.TypeBlock().Accept(v)
}

// VisitTypeLit implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeLit(ctx *cztctl.TypeLitContext) any {
	typeLit := ctx.TypeLitBody().Accept(v)
	alias, ok := typeLit.(*TypeAlias)
	if ok {
		return alias
	}
	doc := v.getDoc(ctx)
	st, ok := typeLit.(*TypeStruct)
	if ok {
		st.DocExpr = doc
		return st
	}
	return typeLit
}

// VisitTypeBlock implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeBlock(ctx *cztctl.TypeBlockContext) any {
	list := ctx.AllTypeBlockBody()
	types := make([]TypeExpr, 0, len(list))
	for _, each := range list {
		types = append(types, each.Accept(v).(TypeExpr))
	}
	return types
}

// VisitTypeLitBody implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeLitBody(ctx *cztctl.TypeLitBodyContext) any {
	if ctx.TypeAlias() != nil {
		return ctx.TypeAlias().Accept(v)
	}
	return ctx.TypeStruct().Accept(v)
}

// VisitTypeBlockBody implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeBlockBody(ctx *cztctl.TypeBlockBodyContext) any {
	if ctx.TypeBlockAlias() != nil {
		return ctx.TypeBlockAlias().Accept(v).(*TypeAlias)
	}
	return ctx.TypeBlockStruct().Accept(v).(*TypeStruct)
}

// VisitTypeStruct implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeStruct(ctx *cztctl.TypeStructContext) any {
	var st TypeStruct
	st.Name = v.newExprWithToken(ctx.GetStructName())
	if ctx.GetStructToken() != nil {
		structExpr := v.newExprWithToken(ctx.GetStructToken())
		structTokenText := ctx.GetStructToken().GetText()
		if structTokenText != "struct" {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', found input '%s'", structTokenText))
		}
		if cztctl.IsGolangKeyWord(structTokenText, "struct") {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', but found golang keyword '%s'", structTokenText))
		}
		st.Struct = structExpr
	}
	st.LBrace = v.newExprWithToken(ctx.GetLbrace())
	st.RBrace = v.newExprWithToken(ctx.GetRbrace())
	fields := ctx.AllField()
	for _, each := range fields {
		f := each.Accept(v)
		if f == nil {
			continue
		}
		st.Fields = append(st.Fields, f.(*TypeField))
	}
	return &st
}

// VisitTypeBlockStruct implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeBlockStruct(ctx *cztctl.TypeBlockStructContext) any {
	var st TypeStruct
	st.Name = v.newExprWithToken(ctx.GetStructName())
	if ctx.GetStructToken() != nil {
		structExpr := v.newExprWithToken(ctx.GetStructToken())
		structTokenText := ctx.GetStructToken().GetText()
		if structTokenText != "struct" {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', found input '%s'", structTokenText))
		}
		if cztctl.IsGolangKeyWord(structTokenText, "struct") {
			v.panic(structExpr, fmt.Sprintf("expecting 'struct', but found golang keyword '%s'", structTokenText))
		}
		st.Struct = structExpr
	}
	st.DocExpr = v.getDoc(ctx)
	st.LBrace = v.newExprWithToken(ctx.GetLbrace())
	st.RBrace = v.newExprWithToken(ctx.GetRbrace())
	fields := ctx.AllField()
	for _, each := range fields {
		f := each.Accept(v)
		if f == nil {
			continue
		}
		st.Fields = append(st.Fields, f.(*TypeField))
	}
	return &st
}

// VisitTypeBlockAlias implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeBlockAlias(ctx *cztctl.TypeBlockAliasContext) any {
	var alias TypeAlias
	alias.Name = v.newExprWithToken(ctx.GetAlias())
	alias.Assign = v.newExprWithToken(ctx.GetAssign())
	alias.DataType = ctx.DataType().Accept(v).(DataType)
	alias.DocExpr = v.getDoc(ctx)
	alias.CommentExpr = v.getComment(ctx)
	v.panic(alias.Name, "unsupported alias")
	return &alias
}

// VisitTypeAlias implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitTypeAlias(ctx *cztctl.TypeAliasContext) any {
	var alias TypeAlias
	alias.Name = v.newExprWithToken(ctx.GetAlias())
	alias.Assign = v.newExprWithToken(ctx.GetAssign())
	alias.DataType = ctx.DataType().Accept(v).(DataType)
	alias.DocExpr = v.getDoc(ctx)
	alias.CommentExpr = v.getComment(ctx)
	v.panic(alias.Name, "unsupported alias")
	return &alias
}

// VisitField implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitField(ctx *cztctl.FieldContext) any {
	iAnonymousFiled := ctx.AnonymousFiled()
	iNormalFieldContext := ctx.NormalField()
	if iAnonymousFiled != nil {
		return iAnonymousFiled.Accept(v).(*TypeField)
	}
	if iNormalFieldContext != nil {
		return iNormalFieldContext.Accept(v).(*TypeField)
	}
	return nil
}

// VisitNormalField implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitNormalField(ctx *cztctl.NormalFieldContext) any {
	var field TypeField
	field.Name = v.newExprWithToken(ctx.GetFieldName())
	iDataTypeContext := ctx.DataType()
	if iDataTypeContext != nil {
		field.DataType = iDataTypeContext.Accept(v).(DataType)
		field.CommentExpr = v.getComment(ctx)
	}
	if ctx.GetTag() != nil {
		tagText := ctx.GetTag().GetText()
		tagExpr := v.newExprWithToken(ctx.GetTag())
		if !cztctl.MatchTag(tagText) {
			v.panic(tagExpr, fmt.Sprintf("mismatched tag, found input '%s'", tagText))
		}
		field.Tag = tagExpr
		field.CommentExpr = v.getComment(ctx)
	}
	field.DocExpr = v.getDoc(ctx)
	return &field
}

// VisitAnonymousFiled implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAnonymousFiled(ctx *cztctl.AnonymousFiledContext) any {
	start := ctx.GetStart()
	stop := ctx.GetStop()
	var field TypeField
	field.IsAnonymous = true
	if ctx.GetStar() != nil {
		nameExpr := v.newExprWithTerminalNode(ctx.ID())
		field.DataType = &Pointer{
			PointerExpr: v.newExprWithText(ctx.GetStar().GetText()+ctx.ID().GetText(), start.GetLine(), start.GetColumn(), start.GetStart(), stop.GetStop()),
			Star:        v.newExprWithToken(ctx.GetStar()),
			Name:        nameExpr,
		}
	} else {
		nameExpr := v.newExprWithTerminalNode(ctx.ID())
		field.DataType = &Literal{Literal: nameExpr}
	}
	field.DocExpr = v.getDoc(ctx)
	field.CommentExpr = v.getComment(ctx)
	return &field
}

// VisitDataType implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitDataType(ctx *cztctl.DataTypeContext) any {
	if ctx.ID() != nil {
		idExpr := v.newExprWithTerminalNode(ctx.ID())
		return &Literal{Literal: idExpr}
	}
	if ctx.MapType() != nil {
		return ctx.MapType().Accept(v)
	}
	if ctx.ArrayType() != nil {
		return ctx.ArrayType().Accept(v)
	}
	if ctx.GetInter() != nil {
		return &Interface{Literal: v.newExprWithToken(ctx.GetInter())}
	}
	if ctx.PointerType() != nil {
		return ctx.PointerType().Accept(v)
	}
	return ctx.TypeStruct().Accept(v)
}

// VisitPointerType implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitPointerType(ctx *cztctl.PointerTypeContext) any {
	nameExpr := v.newExprWithTerminalNode(ctx.ID())
	return &Pointer{
		PointerExpr: v.newExprWithText(ctx.GetText(), ctx.GetStar().GetLine(), ctx.GetStar().GetColumn(), ctx.GetStar().GetStart(), ctx.ID().GetSymbol().GetStop()),
		Star:        v.newExprWithToken(ctx.GetStar()),
		Name:        nameExpr,
	}
}

// VisitMapType implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitMapType(ctx *cztctl.MapTypeContext) any {
	return &Map{
		MapExpr: v.newExprWithText(ctx.GetText(), ctx.GetMapToken().GetLine(), ctx.GetMapToken().GetColumn(),
			ctx.GetMapToken().GetStart(), ctx.GetValue().GetStop().GetStop()),
		Map:    v.newExprWithToken(ctx.GetMapToken()),
		LBrack: v.newExprWithToken(ctx.GetLbrack()),
		RBrack: v.newExprWithToken(ctx.GetRbrack()),
		Key:    v.newExprWithToken(ctx.GetKey()),
		Value:  ctx.GetValue().Accept(v).(DataType),
	}
}

// VisitArrayType implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitArrayType(ctx *cztctl.ArrayTypeContext) any {
	return &Array{
		ArrayExpr: v.newExprWithText(ctx.GetText(), ctx.GetLbrack().GetLine(), ctx.GetLbrack().GetColumn(), ctx.GetLbrack().GetStart(), ctx.DataType().GetStop().GetStop()),
		LBrack:    v.newExprWithToken(ctx.GetLbrack()),
		RBrack:    v.newExprWithToken(ctx.GetRbrack()),
		Literal:   ctx.DataType().Accept(v).(DataType),
	}
}

// NameExpr returns the expression string of TypeAlias
func (a *TypeAlias) NameExpr() Expr { return a.Name }
func (a *TypeAlias) Doc() []Expr    { return a.DocExpr }
func (a *TypeAlias) Comment() Expr  { return a.CommentExpr }
func (a *TypeAlias) Format() error  { return nil }
func (a *TypeAlias) Equal(v any) bool {
	if v == nil {
		return false
	}
	alias := v.(*TypeAlias)
	if !a.Name.Equal(alias.Name) {
		return false
	}
	if !a.Assign.Equal(alias.Assign) {
		return false
	}
	if !a.DataType.Equal(alias.DataType) {
		return false
	}
	return EqualDoc(a, alias)
}

func (l *Literal) Expr() Expr     { return l.Literal }
func (l *Literal) Format() error  { return nil }
func (l *Literal) IsNotNil() bool { return l != nil }
func (l *Literal) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*Literal)
	if !ok {
		return false
	}
	return l.Literal.Equal(v.Literal)
}

func (i *Interface) Expr() Expr     { return i.Literal }
func (i *Interface) Format() error  { return nil }
func (i *Interface) IsNotNil() bool { return i != nil }
func (i *Interface) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*Interface)
	if !ok {
		return false
	}
	return i.Literal.Equal(v.Literal)
}

func (m *Map) Expr() Expr     { return m.MapExpr }
func (m *Map) Format() error  { return nil }
func (m *Map) IsNotNil() bool { return m != nil }
func (m *Map) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*Map)
	if !ok {
		return false
	}
	if !m.Key.Equal(v.Key) {
		return false
	}
	if !m.Value.Equal(v.Value) {
		return false
	}
	if !m.MapExpr.Equal(v.MapExpr) {
		return false
	}
	return m.Map.Equal(v.Map)
}

func (a *Array) Expr() Expr     { return a.ArrayExpr }
func (a *Array) Format() error  { return nil }
func (a *Array) IsNotNil() bool { return a != nil }
func (a *Array) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*Array)
	if !ok {
		return false
	}
	if !a.ArrayExpr.Equal(v.ArrayExpr) {
		return false
	}
	return a.Literal.Equal(v.Literal)
}

func (p *Pointer) Expr() Expr     { return p.PointerExpr }
func (p *Pointer) Format() error  { return nil }
func (p *Pointer) IsNotNil() bool { return p != nil }
func (p *Pointer) Equal(dt DataType) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*Pointer)
	if !ok {
		return false
	}
	if !p.PointerExpr.Equal(v.PointerExpr) {
		return false
	}
	if !p.Star.Equal(v.Star) {
		return false
	}
	return p.Name.Equal(v.Name)
}

func (s *TypeStruct) NameExpr() Expr { return s.Name }
func (s *TypeStruct) Doc() []Expr    { return s.DocExpr }
func (s *TypeStruct) Format() error  { return nil }
func (s *TypeStruct) Equal(dt any) bool {
	if dt == nil {
		return false
	}
	v, ok := dt.(*TypeStruct)
	if !ok {
		return false
	}
	if !s.Name.Equal(v.Name) {
		return false
	}
	var expectDoc, actualDoc []Expr
	expectDoc = append(expectDoc, s.DocExpr...)
	actualDoc = append(actualDoc, v.DocExpr...)
	sort.Slice(expectDoc, func(i, j int) bool { return expectDoc[i].Line() < expectDoc[j].Line() })
	for index, each := range actualDoc {
		if !each.Equal(actualDoc[index]) {
			return false
		}
	}
	if s.Struct != nil {
		if !s.Struct.Equal(v.Struct) {
			return false
		}
	}
	if len(s.Fields) != len(v.Fields) {
		return false
	}
	var expected, actual []*TypeField
	expected = append(expected, s.Fields...)
	actual = append(actual, v.Fields...)
	sort.Slice(expected, func(i, j int) bool { return expected[i].DataType.Expr().Line() < expected[j].DataType.Expr().Line() })
	sort.Slice(actual, func(i, j int) bool { return actual[i].DataType.Expr().Line() < actual[j].DataType.Expr().Line() })
	for index, each := range expected {
		ac := actual[index]
		if !each.Equal(ac) {
			return false
		}
	}
	return true
}

func (t *TypeField) Equal(v any) bool {
	if v == nil {
		return false
	}
	f, ok := v.(*TypeField)
	if !ok {
		return false
	}
	if t.IsAnonymous != f.IsAnonymous {
		return false
	}
	if !t.DataType.Equal(f.DataType) {
		return false
	}
	if !t.IsAnonymous {
		if !t.Name.Equal(f.Name) {
			return false
		}
		if t.Tag != nil {
			if !t.Tag.Equal(f.Tag) {
				return false
			}
		}
	}
	return EqualDoc(t, f)
}

func (t *TypeField) Doc() []Expr   { return t.DocExpr }
func (t *TypeField) Comment() Expr { return t.CommentExpr }
func (t *TypeField) Format() error { return nil }
