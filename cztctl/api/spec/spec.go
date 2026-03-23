package spec

import (
	"errors"
	"strings"
)

type (
	// Doc describes document
	Doc []string

	// Annotation defines key-value
	Annotation struct {
		Properties map[string]string
	}

	// ApiSyntax describes the syntax grammar
	ApiSyntax struct {
		Version string
		Doc     Doc
		Comment Doc
	}

	// ApiSpec describes a parsed extension file (.cron / .rabbitmq)
	ApiSpec struct {
		Info    Info
		Syntax  ApiSyntax
		Imports []Import
		Types   []Type
		Service Service
	}

	// Import describes import statement
	Import struct {
		Value   string
		Doc     Doc
		Comment Doc
	}

	// Group defines a set of routing information
	Group struct {
		Annotation Annotation
		Routes     []Route
	}

	// Info describes info grammar block
	Info struct {
		Title      string
		Desc       string
		Version    string
		Author     string
		Email      string
		Properties map[string]string
	}

	// Member describes the field of a structure
	Member struct {
		Name     string
		Type     Type
		Tag      string
		Comment  string
		Docs     Doc
		IsInline bool
	}

	// Route describes a service route item
	Route struct {
		AtServerAnnotation Annotation
		Method             string // cron: task_type, rabbitmq: queue_name
		Path               string // unused for extension, kept for compatibility
		RequestType        Type   // optional param type
		ResponseType       Type   // unused for extension
		Docs               Doc
		Handler            string
		AtDoc              AtDoc
		HandlerDoc         Doc
		HandlerComment     Doc
		Doc                Doc
		Comment            Doc
		// cron extensions
		Cron      string // cron expression, e.g. "*/1 * * * *"
		CronRetry int    // retry count for asynq.MaxRetry
		// rabbitmq extensions
		Queue string // queue name, e.g. "order.created"
	}

	// Service describes service definition
	Service struct {
		Name   string
		Groups []Group
	}

	// Type defines a type
	Type interface {
		Name() string
		Comments() []string
		Documents() []string
	}

	// DefineStruct describes a structure
	DefineStruct struct {
		RawName string
		Members []Member
		Docs    Doc
	}

	// PrimitiveType describes basic golang types
	PrimitiveType struct {
		RawName string
	}

	// MapType describes a map
	MapType struct {
		RawName string
		Key     string
		Value   Type
	}

	// ArrayType describes a slice
	ArrayType struct {
		RawName string
		Value   Type
	}

	// InterfaceType describes an interface
	InterfaceType struct {
		RawName string
	}

	// PointerType describes a pointer
	PointerType struct {
		RawName string
		Type    Type
	}

	// NestedStruct describes a structure nested in structure
	NestedStruct struct {
		RawName string
		Members []Member
		Docs    Doc
	}

	// AtDoc describes @doc metadata
	AtDoc struct {
		Properties map[string]string
		Text       string
	}

	// Tags represents the parsed struct tags
	Tags struct {
		tags []*Tag
	}

	// Tag represents a single struct tag
	Tag struct {
		Key     string
		Name    string
		Options []string
	}
)

// Name returns the name of DefineStruct
func (d DefineStruct) Name() string { return d.RawName }

// Comments returns nil for DefineStruct
func (d DefineStruct) Comments() []string { return nil }

// Documents returns the documents of DefineStruct
func (d DefineStruct) Documents() []string { return d.Docs }

// Name returns the name of PrimitiveType
func (p PrimitiveType) Name() string { return p.RawName }

// Comments returns nil for PrimitiveType
func (p PrimitiveType) Comments() []string { return nil }

// Documents returns nil for PrimitiveType
func (p PrimitiveType) Documents() []string { return nil }

// Name returns the name of MapType
func (m MapType) Name() string { return m.RawName }

// Comments returns nil for MapType
func (m MapType) Comments() []string { return nil }

// Documents returns nil for MapType
func (m MapType) Documents() []string { return nil }

// Name returns the name of ArrayType
func (a ArrayType) Name() string { return a.RawName }

// Comments returns nil for ArrayType
func (a ArrayType) Comments() []string { return nil }

// Documents returns nil for ArrayType
func (a ArrayType) Documents() []string { return nil }

// Name returns the name of InterfaceType
func (i InterfaceType) Name() string { return i.RawName }

// Comments returns nil for InterfaceType
func (i InterfaceType) Comments() []string { return nil }

// Documents returns nil for InterfaceType
func (i InterfaceType) Documents() []string { return nil }

// Name returns the name of PointerType
func (p PointerType) Name() string { return p.RawName }

// Comments returns nil for PointerType
func (p PointerType) Comments() []string { return nil }

// Documents returns nil for PointerType
func (p PointerType) Documents() []string { return nil }

// Name returns the name of NestedStruct
func (n NestedStruct) Name() string { return n.RawName }

// Comments returns nil for NestedStruct
func (n NestedStruct) Comments() []string { return nil }

// Documents returns the documents of NestedStruct
func (n NestedStruct) Documents() []string { return n.Docs }

// GetComment returns comment value of Member
func (m Member) GetComment() string {
	return strings.TrimSpace(m.Comment)
}

// RequestTypeName returns request type name of route
func (r Route) RequestTypeName() string {
	if r.RequestType == nil {
		return ""
	}
	return r.RequestType.Name()
}

// ResponseTypeName returns response type name of route
func (r Route) ResponseTypeName() string {
	if r.ResponseType == nil {
		return ""
	}
	return r.ResponseType.Name()
}

// GetAnnotation returns the annotation value by key from the group
func (g Group) GetAnnotation(key string) string {
	if g.Annotation.Properties == nil {
		return ""
	}
	return g.Annotation.Properties[key]
}

// GetAnnotation returns the annotation value by key from the route
func (r Route) GetAnnotation(key string) string {
	if r.AtServerAnnotation.Properties == nil {
		return ""
	}
	return r.AtServerAnnotation.Properties[key]
}

// JoinedDoc returns the joined doc string
func (r Route) JoinedDoc() string {
	return strings.Join(r.Doc, " ")
}

// Validate validates the ApiSpec
func (s *ApiSpec) Validate() error {
	if len(s.Service.Groups) == 0 {
		return errors.New("missing service definition")
	}
	return nil
}

// Parse parses a raw tag string into Tags
func Parse(tag string) (*Tags, error) {
	tag = strings.TrimSpace(tag)
	tag = strings.Trim(tag, "`")
	tags := &Tags{}

	for tag != "" {
		// skip leading spaces
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan key
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		key := tag[:i]
		tag = tag[i+1:]

		// scan quoted value
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := tag[1:i]
		tag = tag[i+1:]

		t := &Tag{Key: key}
		parts := strings.Split(qvalue, ",")
		if len(parts) > 0 {
			t.Name = parts[0]
		}
		if len(parts) > 1 {
			t.Options = parts[1:]
		}
		tags.tags = append(tags.tags, t)
	}

	return tags, nil
}

// Get returns the tag by key
func (t *Tags) Get(key string) (*Tag, error) {
	if t == nil {
		return nil, errors.New("tags is nil")
	}
	for _, tag := range t.tags {
		if tag.Key == key {
			return tag, nil
		}
	}
	return nil, errors.New("tag not found: " + key)
}
