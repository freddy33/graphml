package graphml

import (
	"encoding/xml"
	"fmt"
	"io"
)

const (
	// Ext is a file extension for GraphML files.
	Ext = ".graphml"
	// Namespace is a canonical XML namespace for GraphML.
	Namespace = "http://graphml.graphdrawing.org/xmlns"
)

type element interface {
	addAttr(a xml.Attr)
	attrs() []xml.Attr
}

func newAttr(ns, local, value string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: local, Space: ns}, Value: value}
}

// Document is a self-contained GraphML document.
type Document struct {
	Instr  xml.ProcInst
	Attrs  []xml.Attr
	Keys   []Key
	Graphs []Graph `xml:"graph"`
	Data   []Data  `xml:"data"`
}

// Object is a set of common attributes for nodes edges and graphs.
type Object struct {
	ID           string     `xml:"id,attr"`
	Unrecognized []xml.Attr `xml:",any,attr"`
}

func (o *Object) addAttr(a xml.Attr) {
	switch a.Name.Local {
	case "id":
		o.ID = a.Value
	default:
		o.Unrecognized = append(o.Unrecognized, a)
	}
}
func (o *Object) attrs() []xml.Attr {
	out := make([]xml.Attr, 0, len(o.Unrecognized)+1)
	if o.ID != "" {
		out = append(out, newAttr("", "id", o.ID))
	}
	out = append(out, o.Unrecognized...)
	return out
}

// ExtObject is a common set of attributes for nodes that can be extended.
type ExtObject struct {
	Object
	Data []Data `xml:"data"`
}

// NewKey creates a new custom attribute definition.
func NewKey(kind Kind, id, name, typ string) Key {
	return Key{
		Object: Object{
			ID: id,
		},
		For:  kind,
		Name: name, Type: typ,
	}
}

// Create a new Data entry with single token value
func NewData(key string, val interface{}) Data {
	d := Data{}
	d.Key = key
	d.Data = []xml.Token{xml.CharData(fmt.Sprintf("%v", val))}
	return d
}

// Creates a new Edge for given id source and target ids
func NewEdge(id, s, t string) Edge {
	e := Edge{}
	e.ID = id
	e.Source = s
	e.Target = t
	return e
}

// Key is a definition of a custom attribute.
type Key struct {
	Object
	For  Kind   `xml:"for,attr"`
	Name string `xml:"attr.name,attr"`
	Type string `xml:"attr.type,attr"`
}

func (k *Key) addAttr(a xml.Attr) {
	switch a.Name.Local {
	case "for":
		k.For = Kind(a.Value)
	case "attr.name":
		k.Name = a.Value
	case "attr.type":
		k.Type = a.Value
	default:
		k.Object.addAttr(a)
	}
}
func (k *Key) attrs() []xml.Attr {
	attrs := k.Object.attrs()
	attrs = append(attrs, newAttr("", "for", string(k.For)))
	if k.Name != "" {
		attrs = append(attrs, newAttr("", "attr.name", k.Name))
	}
	if k.Type != "" {
		attrs = append(attrs, newAttr("", "attr.type", k.Type))
	}
	return attrs
}

// Graph is a set of nodes and edges.
type Graph struct {
	ExtObject

	// EdgeDefault is a default direction mode for edges (directed or undirected).
	EdgeDefault EdgeDir `xml:"edgedefault,attr"`

	Nodes []Node `xml:"node"`
	Edges []Edge `xml:"edge"`
}

func (g *Graph) addAttr(a xml.Attr) {
	switch a.Name.Local {
	case "edgedefault":
		g.EdgeDefault = EdgeDir(a.Value)
	default:
		g.Object.addAttr(a)
	}
}
func (g *Graph) attrs() []xml.Attr {
	attrs := g.Object.attrs()
	if g.EdgeDefault != "" {
		attrs = append(attrs, newAttr("", "edgedefault", string(g.EdgeDefault)))
	}
	return attrs
}

// Node is a node in a graph.
type Node struct {
	ExtObject

	Graphs []Graph `xml:"graph"`
}

func (n *Node) addAttr(a xml.Attr) {
	n.Object.addAttr(a)
}
func (n *Node) attrs() []xml.Attr {
	return n.Object.attrs()
}

// Edge is a connection between two nodes in a graph.
type Edge struct {
	ExtObject
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

func (e *Edge) addAttr(a xml.Attr) {
	switch a.Name.Local {
	case "source":
		e.Source = a.Value
	case "target":
		e.Target = a.Value
	default:
		e.Object.addAttr(a)
	}
}
func (e *Edge) attrs() []xml.Attr {
	attrs := e.Object.attrs()
	attrs = append(attrs,
		newAttr("", "source", e.Source),
		newAttr("", "target", e.Target),
	)
	return attrs
}

// Data is a raw XML value for a custom attribute.
type Data struct {
	Key          string     `xml:"key,attr"`
	Unrecognized []xml.Attr `xml:",any,attr"`
	Data         []xml.Token
}

// Reader returns a XML token reader for this custom attribute. See xml.NewTokenDecoder().
func (d *Data) Reader() xml.TokenReader {
	return &tokenReader{tokens: d.Data}
}
func (d *Data) addAttr(a xml.Attr) {
	switch a.Name.Local {
	case "key":
		d.Key = a.Value
	default:
		d.Unrecognized = append(d.Unrecognized, a)
	}
}
func (d *Data) attrs() []xml.Attr {
	attrs := make([]xml.Attr, 0, len(d.Unrecognized)+1)
	attrs = append(attrs, newAttr("", "key", d.Key))
	attrs = append(attrs, d.Unrecognized...)
	return attrs
}

type tokenReader struct {
	tokens []xml.Token
}

func (r *tokenReader) Token() (xml.Token, error) {
	if len(r.tokens) == 0 {
		return nil, io.EOF
	}
	t := r.tokens[0]
	r.tokens = r.tokens[1:]
	return t, nil
}

// EdgeDir is a direction mode for edges (directed or undirected).
type EdgeDir string

const (
	EdgeDirected   = EdgeDir("directed")
	EdgeUndirected = EdgeDir("undirected")
)

// Kind is an element kind used for extensions.
type Kind string

const (
	KindAll       = Kind("all")
	KindGraphML   = Kind("graphml")
	KindGraph     = Kind("graph")
	KindNode      = Kind("node")
	KindEdge      = Kind("edge")
	KindHyperEdge = Kind("hyperedge")
	KindPort      = Kind("port")
	KindEndpoint  = Kind("endpoint")
)
