package ast

import (
	"fmt"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/token"
)

// ==================== Core Interfaces ====================

// Node represents a node in the AST.
type Node interface {
	Pos() token.Position
	End() token.Position
	Format(...string) string
	HasHeadCommentGroup() bool
	HasLeadingCommentGroup() bool
	CommentGroup() (head, leading CommentGroup)
}

// Stmt represents a statement in the AST.
type Stmt interface {
	Node
	stmtNode()
}

// Expr represents an expression in the AST.
type Expr interface {
	Node
	exprNode()
}

// AST represents a parsed extension file.
type AST struct {
	Filename string
	Stmts    []Stmt
}

// SyntaxError represents a syntax error.
func SyntaxError(pos token.Position, format string, v ...interface{}) error {
	return fmt.Errorf("syntax error: %s %s", pos.String(), fmt.Sprintf(format, v...))
}

// DuplicateStmtError represents a duplicate statement error.
func DuplicateStmtError(pos token.Position, msg string) error {
	return fmt.Errorf("duplicate declaration: %s %s", pos.String(), msg)
}

// ==================== TokenNode ====================

// TokenNode represents a token node in the AST.
type TokenNode struct {
	HeadCommentGroup    CommentGroup
	Token               token.Token
	LeadingCommentGroup CommentGroup
}

// NewTokenNode creates and returns a new TokenNode.
func NewTokenNode(tok token.Token) *TokenNode {
	return &TokenNode{Token: tok}
}

// RawText returns the node's raw text (stripping quotes).
func (t *TokenNode) RawText() string {
	text := t.Token.Text
	text = strings.TrimPrefix(text, "`")
	text = strings.TrimSuffix(text, "`")
	text = strings.TrimPrefix(text, `"`)
	text = strings.TrimSuffix(text, `"`)
	return text
}

// IsEmptyString returns true if the node is empty string.
func (t *TokenNode) IsEmptyString() bool {
	return t.Token.Text == ""
}

// IsZeroString returns true if the node is zero string.
func (t *TokenNode) IsZeroString() bool {
	return t.Token.Text == `""` || t.Token.Text == "``"
}

// SetLeadingCommentGroup sets the node's leading comment group.
func (t *TokenNode) SetLeadingCommentGroup(cg CommentGroup) {
	t.LeadingCommentGroup = cg
}

func (t *TokenNode) HasLeadingCommentGroup() bool {
	return len(t.LeadingCommentGroup) > 0
}

func (t *TokenNode) HasHeadCommentGroup() bool {
	return len(t.HeadCommentGroup) > 0
}

func (t *TokenNode) CommentGroup() (head, leading CommentGroup) {
	return t.HeadCommentGroup, t.LeadingCommentGroup
}

func (t *TokenNode) PeekFirstLeadingComment() *CommentStmt {
	if len(t.LeadingCommentGroup) > 0 {
		return t.LeadingCommentGroup[0]
	}
	return nil
}

func (t *TokenNode) Format(prefix ...string) string {
	return t.Token.Text
}

func (t *TokenNode) Pos() token.Position {
	if len(t.HeadCommentGroup) > 0 {
		return t.HeadCommentGroup[0].Pos()
	}
	return t.Token.Position
}

func (t *TokenNode) End() token.Position {
	if len(t.LeadingCommentGroup) > 0 {
		return t.LeadingCommentGroup[len(t.LeadingCommentGroup)-1].End()
	}
	return t.Token.Position
}

// ==================== CommentGroup ====================

// CommentGroup represents a list of comments.
type CommentGroup []*CommentStmt

// List returns the list of comment texts.
func (cg CommentGroup) List() []string {
	var list []string
	for _, v := range cg {
		comment := v.Comment.Text
		if strings.TrimSpace(comment) == "" {
			continue
		}
		list = append(list, comment)
	}
	return list
}

// String joins and returns the comment text.
func (cg CommentGroup) String() string {
	return strings.Join(cg.List(), " ")
}

// Valid returns true if the comment group is non-empty.
func (cg CommentGroup) Valid() bool {
	return len(cg) > 0
}

// CommentStmt represents a comment statement.
type CommentStmt struct {
	Comment token.Token
}

func (c *CommentStmt) HasHeadCommentGroup() bool    { return false }
func (c *CommentStmt) HasLeadingCommentGroup() bool { return false }
func (c *CommentStmt) CommentGroup() (head, leading CommentGroup) {
	return
}
func (c *CommentStmt) stmtNode()                      {}
func (c *CommentStmt) Pos() token.Position            { return c.Comment.Position }
func (c *CommentStmt) End() token.Position            { return c.Comment.Position }
func (c *CommentStmt) Format(prefix ...string) string { return c.Comment.Text }

// ==================== KVExpr ====================

// KVExpr is a key value expression.
type KVExpr struct {
	Key   *TokenNode
	Colon *TokenNode
	Value *TokenNode
}

func (i *KVExpr) HasHeadCommentGroup() bool    { return i.Key.HasHeadCommentGroup() }
func (i *KVExpr) HasLeadingCommentGroup() bool { return i.Value.HasLeadingCommentGroup() }
func (i *KVExpr) CommentGroup() (head, leading CommentGroup) {
	return i.Key.HeadCommentGroup, i.Value.LeadingCommentGroup
}
func (i *KVExpr) Format(prefix ...string) string {
	return i.Key.Token.Text + ": " + i.Value.Token.Text
}
func (i *KVExpr) End() token.Position { return i.Value.End() }
func (i *KVExpr) Pos() token.Position { return i.Key.Pos() }
func (i *KVExpr) exprNode()           {}

// ==================== SyntaxStmt ====================

// SyntaxStmt represents syntax = "v1".
type SyntaxStmt struct {
	Syntax *TokenNode
	Assign *TokenNode
	Value  *TokenNode
}

func (s *SyntaxStmt) HasHeadCommentGroup() bool    { return s.Syntax.HasHeadCommentGroup() }
func (s *SyntaxStmt) HasLeadingCommentGroup() bool { return s.Value.HasLeadingCommentGroup() }
func (s *SyntaxStmt) CommentGroup() (head, leading CommentGroup) {
	return s.Syntax.HeadCommentGroup, s.Value.LeadingCommentGroup
}
func (s *SyntaxStmt) Format(prefix ...string) string { return s.Value.Token.Text }
func (s *SyntaxStmt) End() token.Position            { return s.Value.End() }
func (s *SyntaxStmt) Pos() token.Position            { return s.Syntax.Pos() }
func (s *SyntaxStmt) stmtNode()                      {}

// ==================== InfoStmt ====================

// InfoStmt represents the info block.
type InfoStmt struct {
	Info   *TokenNode
	LParen *TokenNode
	Values []*KVExpr
	RParen *TokenNode
}

func (i *InfoStmt) HasHeadCommentGroup() bool    { return i.Info.HasHeadCommentGroup() }
func (i *InfoStmt) HasLeadingCommentGroup() bool { return i.RParen.HasLeadingCommentGroup() }
func (i *InfoStmt) CommentGroup() (head, leading CommentGroup) {
	return i.Info.HeadCommentGroup, i.RParen.LeadingCommentGroup
}
func (i *InfoStmt) Format(prefix ...string) string { return "info(...)" }
func (i *InfoStmt) End() token.Position            { return i.RParen.End() }
func (i *InfoStmt) Pos() token.Position            { return i.Info.Pos() }
func (i *InfoStmt) stmtNode()                      {}

// ==================== ImportStmt ====================

// ImportStmt represents an import statement.
type ImportStmt interface {
	Stmt
	importNode()
}

// ImportLiteralStmt represents import "file".
type ImportLiteralStmt struct {
	Import *TokenNode
	Value  *TokenNode
}

func (i *ImportLiteralStmt) HasHeadCommentGroup() bool    { return i.Import.HasHeadCommentGroup() }
func (i *ImportLiteralStmt) HasLeadingCommentGroup() bool { return i.Value.HasLeadingCommentGroup() }
func (i *ImportLiteralStmt) CommentGroup() (head, leading CommentGroup) {
	return i.Import.HeadCommentGroup, i.Value.LeadingCommentGroup
}
func (i *ImportLiteralStmt) Format(prefix ...string) string { return i.Value.Token.Text }
func (i *ImportLiteralStmt) End() token.Position            { return i.Value.End() }
func (i *ImportLiteralStmt) importNode()                    {}
func (i *ImportLiteralStmt) Pos() token.Position            { return i.Import.Pos() }
func (i *ImportLiteralStmt) stmtNode()                      {}

// ImportGroupStmt represents import (...).
type ImportGroupStmt struct {
	Import *TokenNode
	LParen *TokenNode
	Values []*TokenNode
	RParen *TokenNode
}

func (i *ImportGroupStmt) HasHeadCommentGroup() bool    { return i.Import.HasHeadCommentGroup() }
func (i *ImportGroupStmt) HasLeadingCommentGroup() bool { return i.RParen.HasLeadingCommentGroup() }
func (i *ImportGroupStmt) CommentGroup() (head, leading CommentGroup) {
	return i.Import.HeadCommentGroup, i.RParen.LeadingCommentGroup
}
func (i *ImportGroupStmt) Format(prefix ...string) string { return "import(...)" }
func (i *ImportGroupStmt) End() token.Position            { return i.RParen.End() }
func (i *ImportGroupStmt) importNode()                    {}
func (i *ImportGroupStmt) Pos() token.Position            { return i.Import.Pos() }
func (i *ImportGroupStmt) stmtNode()                      {}

// ==================== TypeStmt ====================

// TypeStmt is the interface for type statement.
type TypeStmt interface {
	Stmt
	typeNode()
}

// TypeLiteralStmt represents type Name struct{...}.
type TypeLiteralStmt struct {
	Type *TokenNode
	Expr *TypeExpr
}

func (t *TypeLiteralStmt) HasHeadCommentGroup() bool    { return t.Type.HasHeadCommentGroup() }
func (t *TypeLiteralStmt) HasLeadingCommentGroup() bool { return t.Expr.HasLeadingCommentGroup() }
func (t *TypeLiteralStmt) CommentGroup() (head, leading CommentGroup) {
	_, leading = t.Expr.CommentGroup()
	return t.Type.HeadCommentGroup, leading
}
func (t *TypeLiteralStmt) Format(prefix ...string) string { return "type " + t.Expr.Name.Token.Text }
func (t *TypeLiteralStmt) End() token.Position            { return t.Expr.End() }
func (t *TypeLiteralStmt) Pos() token.Position            { return t.Type.Pos() }
func (t *TypeLiteralStmt) stmtNode()                      {}
func (t *TypeLiteralStmt) typeNode()                      {}

// TypeGroupStmt represents type (...).
type TypeGroupStmt struct {
	Type     *TokenNode
	LParen   *TokenNode
	ExprList []*TypeExpr
	RParen   *TokenNode
}

func (t *TypeGroupStmt) HasHeadCommentGroup() bool    { return t.Type.HasHeadCommentGroup() }
func (t *TypeGroupStmt) HasLeadingCommentGroup() bool { return t.RParen.HasLeadingCommentGroup() }
func (t *TypeGroupStmt) CommentGroup() (head, leading CommentGroup) {
	return t.Type.HeadCommentGroup, t.RParen.LeadingCommentGroup
}
func (t *TypeGroupStmt) Format(prefix ...string) string { return "type(...)" }
func (t *TypeGroupStmt) End() token.Position            { return t.RParen.End() }
func (t *TypeGroupStmt) Pos() token.Position            { return t.Type.Pos() }
func (t *TypeGroupStmt) stmtNode()                      {}
func (t *TypeGroupStmt) typeNode()                      {}

// TypeExpr represents a single type expression.
type TypeExpr struct {
	Name     *TokenNode
	Assign   *TokenNode
	DataType DataType
}

func (e *TypeExpr) HasHeadCommentGroup() bool    { return e.Name.HasHeadCommentGroup() }
func (e *TypeExpr) HasLeadingCommentGroup() bool { return e.DataType.HasLeadingCommentGroup() }
func (e *TypeExpr) CommentGroup() (head, leading CommentGroup) {
	_, leading = e.DataType.CommentGroup()
	return e.Name.HeadCommentGroup, leading
}
func (e *TypeExpr) Format(prefix ...string) string { return e.Name.Token.Text }
func (e *TypeExpr) End() token.Position            { return e.DataType.End() }
func (e *TypeExpr) Pos() token.Position            { return e.Name.Pos() }
func (e *TypeExpr) exprNode()                      {}

// ==================== DataType ====================

// DataType represents a data type.
type DataType interface {
	Expr
	dataTypeNode()
	CanEqual() bool
	ContainsStruct() bool
	RawText() string
}

// BaseDataType is a base data type (int, string, etc) or a user-defined type name.
type BaseDataType struct {
	Base *TokenNode
}

func (t *BaseDataType) HasHeadCommentGroup() bool    { return t.Base.HasHeadCommentGroup() }
func (t *BaseDataType) HasLeadingCommentGroup() bool { return t.Base.HasLeadingCommentGroup() }
func (t *BaseDataType) CommentGroup() (head, leading CommentGroup) {
	return t.Base.HeadCommentGroup, t.Base.LeadingCommentGroup
}
func (t *BaseDataType) Format(prefix ...string) string { return t.Base.Token.Text }
func (t *BaseDataType) End() token.Position            { return t.Base.End() }
func (t *BaseDataType) RawText() string                { return t.Base.Token.Text }
func (t *BaseDataType) ContainsStruct() bool           { return false }
func (t *BaseDataType) CanEqual() bool                 { return true }
func (t *BaseDataType) Pos() token.Position            { return t.Base.Pos() }
func (t *BaseDataType) exprNode()                      {}
func (t *BaseDataType) dataTypeNode()                  {}

// AnyDataType is the any / interface{} type.
type AnyDataType struct {
	Any *TokenNode
}

func (t *AnyDataType) HasHeadCommentGroup() bool    { return t.Any.HasHeadCommentGroup() }
func (t *AnyDataType) HasLeadingCommentGroup() bool { return t.Any.HasLeadingCommentGroup() }
func (t *AnyDataType) CommentGroup() (head, leading CommentGroup) {
	return t.Any.HeadCommentGroup, t.Any.LeadingCommentGroup
}
func (t *AnyDataType) Format(prefix ...string) string { return t.Any.Token.Text }
func (t *AnyDataType) End() token.Position            { return t.Any.End() }
func (t *AnyDataType) RawText() string                { return t.Any.Token.Text }
func (t *AnyDataType) ContainsStruct() bool           { return false }
func (t *AnyDataType) CanEqual() bool                 { return true }
func (t *AnyDataType) Pos() token.Position            { return t.Any.Pos() }
func (t *AnyDataType) exprNode()                      {}
func (t *AnyDataType) dataTypeNode()                  {}

// InterfaceDataType is the interface{} type.
type InterfaceDataType struct {
	Interface *TokenNode
}

func (t *InterfaceDataType) HasHeadCommentGroup() bool { return t.Interface.HasHeadCommentGroup() }
func (t *InterfaceDataType) HasLeadingCommentGroup() bool {
	return t.Interface.HasLeadingCommentGroup()
}
func (t *InterfaceDataType) CommentGroup() (head, leading CommentGroup) {
	return t.Interface.HeadCommentGroup, t.Interface.LeadingCommentGroup
}
func (t *InterfaceDataType) Format(prefix ...string) string { return t.Interface.Token.Text }
func (t *InterfaceDataType) End() token.Position            { return t.Interface.End() }
func (t *InterfaceDataType) RawText() string                { return t.Interface.Token.Text }
func (t *InterfaceDataType) ContainsStruct() bool           { return false }
func (t *InterfaceDataType) CanEqual() bool                 { return true }
func (t *InterfaceDataType) Pos() token.Position            { return t.Interface.Pos() }
func (t *InterfaceDataType) exprNode()                      {}
func (t *InterfaceDataType) dataTypeNode()                  {}

// StructDataType is a struct { ... } type.
type StructDataType struct {
	LBrace   *TokenNode
	Elements ElemExprList
	RBrace   *TokenNode
}

func (t *StructDataType) HasHeadCommentGroup() bool    { return t.LBrace.HasHeadCommentGroup() }
func (t *StructDataType) HasLeadingCommentGroup() bool { return t.RBrace.HasLeadingCommentGroup() }
func (t *StructDataType) CommentGroup() (head, leading CommentGroup) {
	return t.LBrace.HeadCommentGroup, t.RBrace.LeadingCommentGroup
}
func (t *StructDataType) Format(prefix ...string) string {
	// Simplified: return "{...}" for raw text extraction
	var parts []string
	parts = append(parts, "{")
	for _, elem := range t.Elements {
		var names []string
		for _, n := range elem.Name {
			names = append(names, n.Token.Text)
		}
		line := strings.Join(names, ", ")
		if elem.DataType != nil {
			line += " " + elem.DataType.RawText()
		}
		if elem.Tag != nil {
			line += " " + elem.Tag.Token.Text
		}
		parts = append(parts, line)
	}
	parts = append(parts, "}")
	return strings.Join(parts, "\n")
}
func (t *StructDataType) End() token.Position  { return t.RBrace.End() }
func (t *StructDataType) RawText() string      { return t.Format("") }
func (t *StructDataType) ContainsStruct() bool { return true }
func (t *StructDataType) CanEqual() bool {
	for _, v := range t.Elements {
		if !v.DataType.CanEqual() {
			return false
		}
	}
	return true
}
func (t *StructDataType) Pos() token.Position { return t.LBrace.Pos() }
func (t *StructDataType) exprNode()           {}
func (t *StructDataType) dataTypeNode()       {}

// MapDataType is a map[K]V type.
type MapDataType struct {
	Map    *TokenNode
	LBrack *TokenNode
	Key    DataType
	RBrack *TokenNode
	Value  DataType
}

func (t *MapDataType) HasHeadCommentGroup() bool    { return t.Map.HasHeadCommentGroup() }
func (t *MapDataType) HasLeadingCommentGroup() bool { return t.Value.HasLeadingCommentGroup() }
func (t *MapDataType) CommentGroup() (head, leading CommentGroup) {
	_, leading = t.Value.CommentGroup()
	return t.Map.HeadCommentGroup, leading
}
func (t *MapDataType) Format(prefix ...string) string {
	return "map[" + t.Key.RawText() + "]" + t.Value.RawText()
}
func (t *MapDataType) End() token.Position { return t.Value.End() }
func (t *MapDataType) RawText() string     { return t.Format("") }
func (t *MapDataType) ContainsStruct() bool {
	return t.Key.ContainsStruct() || t.Value.ContainsStruct()
}
func (t *MapDataType) CanEqual() bool      { return false }
func (t *MapDataType) Pos() token.Position { return t.Map.Pos() }
func (t *MapDataType) exprNode()           {}
func (t *MapDataType) dataTypeNode()       {}

// PointerDataType is a *T type.
type PointerDataType struct {
	Star     *TokenNode
	DataType DataType
}

func (t *PointerDataType) HasHeadCommentGroup() bool    { return t.Star.HasHeadCommentGroup() }
func (t *PointerDataType) HasLeadingCommentGroup() bool { return t.DataType.HasLeadingCommentGroup() }
func (t *PointerDataType) CommentGroup() (head, leading CommentGroup) {
	_, leading = t.DataType.CommentGroup()
	return t.Star.HeadCommentGroup, leading
}
func (t *PointerDataType) Format(prefix ...string) string { return "*" + t.DataType.RawText() }
func (t *PointerDataType) End() token.Position            { return t.DataType.End() }
func (t *PointerDataType) RawText() string                { return t.Format("") }
func (t *PointerDataType) ContainsStruct() bool           { return t.DataType.ContainsStruct() }
func (t *PointerDataType) CanEqual() bool                 { return t.DataType.CanEqual() }
func (t *PointerDataType) Pos() token.Position            { return t.Star.Pos() }
func (t *PointerDataType) exprNode()                      {}
func (t *PointerDataType) dataTypeNode()                  {}

// SliceDataType is a []T type.
type SliceDataType struct {
	LBrack   *TokenNode
	RBrack   *TokenNode
	DataType DataType
}

func (t *SliceDataType) HasHeadCommentGroup() bool    { return t.LBrack.HasHeadCommentGroup() }
func (t *SliceDataType) HasLeadingCommentGroup() bool { return t.DataType.HasLeadingCommentGroup() }
func (t *SliceDataType) CommentGroup() (head, leading CommentGroup) {
	_, leading = t.DataType.CommentGroup()
	return t.LBrack.HeadCommentGroup, leading
}
func (t *SliceDataType) Format(prefix ...string) string { return "[]" + t.DataType.RawText() }
func (t *SliceDataType) End() token.Position            { return t.DataType.End() }
func (t *SliceDataType) RawText() string                { return t.Format("") }
func (t *SliceDataType) ContainsStruct() bool           { return t.DataType.ContainsStruct() }
func (t *SliceDataType) CanEqual() bool                 { return false }
func (t *SliceDataType) Pos() token.Position            { return t.LBrack.Pos() }
func (t *SliceDataType) exprNode()                      {}
func (t *SliceDataType) dataTypeNode()                  {}

// ArrayDataType is a [N]T type.
type ArrayDataType struct {
	LBrack   *TokenNode
	Length   *TokenNode
	RBrack   *TokenNode
	DataType DataType
}

func (t *ArrayDataType) HasHeadCommentGroup() bool    { return t.LBrack.HasHeadCommentGroup() }
func (t *ArrayDataType) HasLeadingCommentGroup() bool { return t.DataType.HasLeadingCommentGroup() }
func (t *ArrayDataType) CommentGroup() (head, leading CommentGroup) {
	_, leading = t.DataType.CommentGroup()
	return t.LBrack.HeadCommentGroup, leading
}
func (t *ArrayDataType) Format(prefix ...string) string {
	return "[" + t.Length.Token.Text + "]" + t.DataType.RawText()
}
func (t *ArrayDataType) End() token.Position  { return t.DataType.End() }
func (t *ArrayDataType) RawText() string      { return t.Format("") }
func (t *ArrayDataType) ContainsStruct() bool { return t.DataType.ContainsStruct() }
func (t *ArrayDataType) CanEqual() bool       { return t.DataType.CanEqual() }
func (t *ArrayDataType) Pos() token.Position  { return t.LBrack.Pos() }
func (t *ArrayDataType) exprNode()            {}
func (t *ArrayDataType) dataTypeNode()        {}

// ==================== ElemExpr ====================

// ElemExpr represents a struct field.
type ElemExpr struct {
	Name     []*TokenNode
	DataType DataType
	Tag      *TokenNode
}

// IsAnonymous returns true if the element is anonymous (embedded).
func (e *ElemExpr) IsAnonymous() bool {
	return len(e.Name) == 0
}

func (e *ElemExpr) HasHeadCommentGroup() bool {
	if e.IsAnonymous() {
		return e.DataType.HasHeadCommentGroup()
	}
	return e.Name[0].HasHeadCommentGroup()
}

func (e *ElemExpr) HasLeadingCommentGroup() bool {
	if e.Tag != nil {
		return e.Tag.HasLeadingCommentGroup()
	}
	return e.DataType.HasLeadingCommentGroup()
}

func (e *ElemExpr) CommentGroup() (head, leading CommentGroup) {
	if e.Tag != nil {
		leading = e.Tag.LeadingCommentGroup
	} else {
		_, leading = e.DataType.CommentGroup()
	}
	if e.IsAnonymous() {
		head, _ = e.DataType.CommentGroup()
		return head, leading
	}
	return e.Name[0].HeadCommentGroup, leading
}

func (e *ElemExpr) Format(prefix ...string) string {
	var parts []string
	for _, n := range e.Name {
		parts = append(parts, n.Token.Text)
	}
	return strings.Join(parts, ", ")
}

func (e *ElemExpr) End() token.Position {
	if e.Tag != nil {
		return e.Tag.End()
	}
	return e.DataType.End()
}

func (e *ElemExpr) Pos() token.Position {
	if len(e.Name) > 0 {
		return e.Name[0].Pos()
	}
	return token.IllegalPosition
}

func (e *ElemExpr) exprNode() {}

// ElemExprList is a list of struct fields.
type ElemExprList []*ElemExpr

// ==================== Service Nodes ====================

// AtServerStmt represents @server (...) block.
type AtServerStmt struct {
	AtServer *TokenNode
	LParen   *TokenNode
	Values   []*KVExpr
	RParen   *TokenNode
}

func (a *AtServerStmt) HasHeadCommentGroup() bool    { return a.AtServer.HasHeadCommentGroup() }
func (a *AtServerStmt) HasLeadingCommentGroup() bool { return a.RParen.HasLeadingCommentGroup() }
func (a *AtServerStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtServer.HeadCommentGroup, a.RParen.LeadingCommentGroup
}
func (a *AtServerStmt) Format(prefix ...string) string { return "@server(...)" }
func (a *AtServerStmt) End() token.Position            { return a.RParen.End() }
func (a *AtServerStmt) Pos() token.Position            { return a.AtServer.Pos() }
func (a *AtServerStmt) stmtNode()                      {}

// AtDocStmt is the interface for @doc statements.
type AtDocStmt interface {
	Stmt
	atDocNode()
}

// AtDocLiteralStmt represents @doc "string".
type AtDocLiteralStmt struct {
	AtDoc *TokenNode
	Value *TokenNode
}

func (a *AtDocLiteralStmt) HasHeadCommentGroup() bool    { return a.AtDoc.HasHeadCommentGroup() }
func (a *AtDocLiteralStmt) HasLeadingCommentGroup() bool { return a.Value.HasLeadingCommentGroup() }
func (a *AtDocLiteralStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtDoc.HeadCommentGroup, a.Value.LeadingCommentGroup
}
func (a *AtDocLiteralStmt) Format(prefix ...string) string { return a.Value.Token.Text }
func (a *AtDocLiteralStmt) End() token.Position            { return a.Value.End() }
func (a *AtDocLiteralStmt) atDocNode()                     {}
func (a *AtDocLiteralStmt) Pos() token.Position            { return a.AtDoc.Pos() }
func (a *AtDocLiteralStmt) stmtNode()                      {}

// AtDocGroupStmt represents @doc(key: value ...).
type AtDocGroupStmt struct {
	AtDoc  *TokenNode
	LParen *TokenNode
	Values []*KVExpr
	RParen *TokenNode
}

func (a *AtDocGroupStmt) HasHeadCommentGroup() bool    { return a.AtDoc.HasHeadCommentGroup() }
func (a *AtDocGroupStmt) HasLeadingCommentGroup() bool { return a.RParen.HasLeadingCommentGroup() }
func (a *AtDocGroupStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtDoc.HeadCommentGroup, a.RParen.LeadingCommentGroup
}
func (a *AtDocGroupStmt) Format(prefix ...string) string { return "@doc(...)" }
func (a *AtDocGroupStmt) End() token.Position            { return a.RParen.End() }
func (a *AtDocGroupStmt) atDocNode()                     {}
func (a *AtDocGroupStmt) Pos() token.Position            { return a.AtDoc.Pos() }
func (a *AtDocGroupStmt) stmtNode()                      {}

// AtHandlerStmt represents @handler Name.
type AtHandlerStmt struct {
	AtHandler *TokenNode
	Name      *TokenNode
}

func (a *AtHandlerStmt) HasHeadCommentGroup() bool    { return a.AtHandler.HasHeadCommentGroup() }
func (a *AtHandlerStmt) HasLeadingCommentGroup() bool { return a.Name.HasLeadingCommentGroup() }
func (a *AtHandlerStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtHandler.HeadCommentGroup, a.Name.LeadingCommentGroup
}
func (a *AtHandlerStmt) Format(prefix ...string) string { return "@handler " + a.Name.Token.Text }
func (a *AtHandlerStmt) End() token.Position            { return a.Name.End() }
func (a *AtHandlerStmt) Pos() token.Position            { return a.AtHandler.Pos() }
func (a *AtHandlerStmt) stmtNode()                      {}

// AtCronStmt represents @cron "expression".
type AtCronStmt struct {
	AtCron *TokenNode
	Value  *TokenNode
}

func (a *AtCronStmt) HasHeadCommentGroup() bool    { return a.AtCron.HasHeadCommentGroup() }
func (a *AtCronStmt) HasLeadingCommentGroup() bool { return a.Value.HasLeadingCommentGroup() }
func (a *AtCronStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtCron.HeadCommentGroup, a.Value.LeadingCommentGroup
}
func (a *AtCronStmt) Format(prefix ...string) string { return "@cron " + a.Value.Token.Text }
func (a *AtCronStmt) End() token.Position            { return a.Value.End() }
func (a *AtCronStmt) Pos() token.Position            { return a.AtCron.Pos() }
func (a *AtCronStmt) stmtNode()                      {}

// AtCronRetryStmt represents @cronRetry N.
type AtCronRetryStmt struct {
	AtCronRetry *TokenNode
	Value       *TokenNode
}

func (a *AtCronRetryStmt) HasHeadCommentGroup() bool    { return a.AtCronRetry.HasHeadCommentGroup() }
func (a *AtCronRetryStmt) HasLeadingCommentGroup() bool { return a.Value.HasLeadingCommentGroup() }
func (a *AtCronRetryStmt) CommentGroup() (head, leading CommentGroup) {
	return a.AtCronRetry.HeadCommentGroup, a.Value.LeadingCommentGroup
}
func (a *AtCronRetryStmt) Format(prefix ...string) string { return "@cronRetry " + a.Value.Token.Text }
func (a *AtCronRetryStmt) End() token.Position            { return a.Value.End() }
func (a *AtCronRetryStmt) Pos() token.Position            { return a.AtCronRetry.Pos() }
func (a *AtCronRetryStmt) stmtNode()                      {}

// ==================== ServiceStmt ====================

// ServiceStmt represents the service block.
type ServiceStmt struct {
	AtServerStmt *AtServerStmt
	Service      *TokenNode
	Name         *ServiceNameExpr
	LBrace       *TokenNode
	Routes       []*ServiceItemStmt
	RBrace       *TokenNode
}

func (s *ServiceStmt) HasHeadCommentGroup() bool {
	if s.AtServerStmt != nil {
		return s.AtServerStmt.HasHeadCommentGroup()
	}
	return s.Service.HasHeadCommentGroup()
}
func (s *ServiceStmt) HasLeadingCommentGroup() bool { return s.RBrace.HasLeadingCommentGroup() }
func (s *ServiceStmt) CommentGroup() (head, leading CommentGroup) {
	if s.AtServerStmt != nil {
		head, _ = s.AtServerStmt.CommentGroup()
		return head, s.RBrace.LeadingCommentGroup
	}
	return s.Service.HeadCommentGroup, s.RBrace.LeadingCommentGroup
}
func (s *ServiceStmt) Format(prefix ...string) string {
	return "service " + s.Name.Format()
}
func (s *ServiceStmt) End() token.Position { return s.RBrace.End() }
func (s *ServiceStmt) Pos() token.Position {
	if s.AtServerStmt != nil {
		return s.AtServerStmt.Pos()
	}
	return s.Service.Pos()
}
func (s *ServiceStmt) stmtNode() {}

// ServiceNameExpr represents a service name (may include -api suffix).
type ServiceNameExpr struct {
	Name *TokenNode
}

func (s *ServiceNameExpr) HasHeadCommentGroup() bool    { return s.Name.HasHeadCommentGroup() }
func (s *ServiceNameExpr) HasLeadingCommentGroup() bool { return s.Name.HasLeadingCommentGroup() }
func (s *ServiceNameExpr) CommentGroup() (head, leading CommentGroup) {
	return s.Name.HeadCommentGroup, s.Name.LeadingCommentGroup
}
func (s *ServiceNameExpr) Format(prefix ...string) string { return s.Name.Token.Text }
func (s *ServiceNameExpr) End() token.Position            { return s.Name.End() }
func (s *ServiceNameExpr) Pos() token.Position            { return s.Name.Pos() }
func (s *ServiceNameExpr) exprNode()                      {}

// ServiceItemStmt represents a single route item in service block.
type ServiceItemStmt struct {
	AtDoc       AtDocStmt
	AtCron      *AtCronStmt
	AtCronRetry *AtCronRetryStmt
	AtHandler   *AtHandlerStmt
	Route       *RouteStmt
}

func (s *ServiceItemStmt) HasHeadCommentGroup() bool {
	if s.AtDoc != nil {
		return s.AtDoc.HasHeadCommentGroup()
	}
	if s.AtCron != nil {
		return s.AtCron.HasHeadCommentGroup()
	}
	return s.AtHandler.HasHeadCommentGroup()
}
func (s *ServiceItemStmt) HasLeadingCommentGroup() bool {
	return s.Route.HasLeadingCommentGroup()
}
func (s *ServiceItemStmt) CommentGroup() (head, leading CommentGroup) {
	_, leading = s.Route.CommentGroup()
	if s.AtDoc != nil {
		head, _ = s.AtDoc.CommentGroup()
		return head, leading
	}
	if s.AtCron != nil {
		head, _ = s.AtCron.CommentGroup()
		return head, leading
	}
	head, _ = s.AtHandler.CommentGroup()
	return head, leading
}
func (s *ServiceItemStmt) Format(prefix ...string) string {
	return s.Route.Format()
}
func (s *ServiceItemStmt) End() token.Position { return s.Route.End() }
func (s *ServiceItemStmt) Pos() token.Position {
	if s.AtDoc != nil {
		return s.AtDoc.Pos()
	}
	if s.AtCron != nil {
		return s.AtCron.Pos()
	}
	return s.AtHandler.Pos()
}
func (s *ServiceItemStmt) stmtNode() {}

// ==================== RouteStmt (Extension) ====================

// RouteStmt represents a route: TaskType(Param) or queue.name
type RouteStmt struct {
	// Name is the task_type or queue_name token (already joined for dot-separated names).
	Name *TokenNode
	// Request is the optional parameter type: (ParamType).
	Request *BodyStmt
}

func (r *RouteStmt) HasHeadCommentGroup() bool { return r.Name.HasHeadCommentGroup() }
func (r *RouteStmt) HasLeadingCommentGroup() bool {
	if r.Request != nil {
		return r.Request.HasLeadingCommentGroup()
	}
	return r.Name.HasLeadingCommentGroup()
}
func (r *RouteStmt) CommentGroup() (head, leading CommentGroup) {
	head, _ = r.Name.CommentGroup()
	if r.Request != nil {
		_, leading = r.Request.CommentGroup()
	} else {
		_, leading = r.Name.CommentGroup()
	}
	return head, leading
}
func (r *RouteStmt) Format(prefix ...string) string {
	text := r.Name.Token.Text
	if r.Request != nil && r.Request.Body != nil {
		text += "(" + r.Request.Body.Value.Token.Text + ")"
	}
	return text
}
func (r *RouteStmt) End() token.Position {
	if r.Request != nil {
		return r.Request.End()
	}
	return r.Name.End()
}
func (r *RouteStmt) Pos() token.Position { return r.Name.Pos() }
func (r *RouteStmt) stmtNode()           {}

// ==================== BodyStmt / BodyExpr ====================

// BodyStmt represents (ParamType).
type BodyStmt struct {
	LParen *TokenNode
	Body   *BodyExpr
	RParen *TokenNode
}

func (b *BodyStmt) HasHeadCommentGroup() bool    { return b.LParen.HasHeadCommentGroup() }
func (b *BodyStmt) HasLeadingCommentGroup() bool { return b.RParen.HasLeadingCommentGroup() }
func (b *BodyStmt) CommentGroup() (head, leading CommentGroup) {
	return b.LParen.HeadCommentGroup, b.RParen.LeadingCommentGroup
}
func (b *BodyStmt) Format(prefix ...string) string {
	if b.Body == nil {
		return "()"
	}
	return "(" + b.Body.Format() + ")"
}
func (b *BodyStmt) End() token.Position { return b.RParen.End() }
func (b *BodyStmt) Pos() token.Position { return b.LParen.Pos() }
func (b *BodyStmt) stmtNode()           {}

// BodyExpr represents the content inside parentheses.
type BodyExpr struct {
	LBrack *TokenNode // for []Type
	RBrack *TokenNode
	Star   *TokenNode // for *Type
	Value  *TokenNode // type name
}

func (e *BodyExpr) HasHeadCommentGroup() bool {
	if e.LBrack != nil {
		return e.LBrack.HasHeadCommentGroup()
	}
	if e.Star != nil {
		return e.Star.HasHeadCommentGroup()
	}
	return e.Value.HasHeadCommentGroup()
}
func (e *BodyExpr) HasLeadingCommentGroup() bool {
	return e.Value.HasLeadingCommentGroup()
}
func (e *BodyExpr) CommentGroup() (head, leading CommentGroup) {
	if e.LBrack != nil {
		head = e.LBrack.HeadCommentGroup
	} else if e.Star != nil {
		head = e.Star.HeadCommentGroup
	} else {
		head = e.Value.HeadCommentGroup
	}
	return head, e.Value.LeadingCommentGroup
}
func (e *BodyExpr) Format(prefix ...string) string {
	var text string
	if e.LBrack != nil {
		text += "[]"
	}
	if e.Star != nil {
		text += "*"
	}
	text += e.Value.Token.Text
	return text
}
func (e *BodyExpr) End() token.Position { return e.Value.End() }
func (e *BodyExpr) Pos() token.Position {
	if e.LBrack != nil {
		return e.LBrack.Pos()
	}
	if e.Star != nil {
		return e.Star.Pos()
	}
	return e.Value.Pos()
}
func (e *BodyExpr) exprNode() {}

// IsArrayType returns true if the body expr is an array type.
func (e *BodyExpr) IsArrayType() bool {
	return e.LBrack != nil
}
