package ir

import (
	"errors"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/multi"
	"strings"
)

type dotNode struct {
	id    int64
	ports []string
	name  string
}

func (n dotNode) ID() int64 { return n.id }
func (n dotNode) Attributes() []encoding.Attribute {
	var ports = []string{}
	for _, p := range n.ports {
		ports = append(ports, fmt.Sprintf("<%s> %s", p, p))
	}
	label := fmt.Sprintf("{%s | { %s }}", n.name, strings.Join(ports, " | "))
	attrs := []encoding.Attribute{
		{Key: "shape", Value: "record"},
		{Key: "label", Value: label},
	}
	return attrs
}

type dotPortedEdge struct {
	id          int64
	from, to    graph.Node
	fromPort    string
	fromCompass string
	toPort      string
	toCompass   string
}

func (e dotPortedEdge) From() graph.Node { return e.from }
func (e dotPortedEdge) To() graph.Node   { return e.to }
func (e dotPortedEdge) ID() int64        { return e.id }
func (e dotPortedEdge) ReversedEdge() graph.Edge {
	e.from, e.to = e.to, e.from
	e.fromPort, e.toPort = e.toPort, e.fromPort
	e.fromCompass, e.toCompass = e.toCompass, e.fromCompass
	return e
}
func (e dotPortedEdge) ReversedLine() graph.Line {
	e.from, e.to = e.to, e.from
	e.fromPort, e.toPort = e.toPort, e.fromPort
	e.fromCompass, e.toCompass = e.toCompass, e.fromCompass
	return e
}

func (e dotPortedEdge) Weight() float64 { return 0 }

func (e dotPortedEdge) FromPort() (port, compass string) {
	return e.fromPort, e.fromCompass
}
func (e dotPortedEdge) ToPort() (port, compass string) {
	return e.toPort, e.toCompass
}

type dotBuilder struct {
	nodeCounter int64
	edgeCounter int64
	nodes       []*dotNode
	edges       []*dotPortedEdge
	typeCache   map[Type]*dotNode
	dataCache   map[Data]*dotNode
	kindCache   map[Kind]*dotNode
}

func dotWalkUnit(b *dotBuilder, u Unit) graph.Node {
	switch x := u.(type) {
	//case *EnumType:
	//case *Load:
	case *Save:
		slot_field := fmt.Sprintf("Slot: %s", x.Slot)
		n := &dotNode{
			b.getNodeId(),
			[]string{slot_field, "Path", "Data"},
			"Save",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Path {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkData(b, a),
				fromPort: "Path"}
			b.edges = append(b.edges, &e)

		}
		e := &dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.Data),
			fromPort: "Data"}
		b.edges = append(b.edges, e)
		return n
		//Path []Data
		//Data Data
	//case *Emit:
	//case *Send:
	//case *Have:
	//case *AbsTD:
	//case *AbsDD:
	//case *AppDD:
	//case *AppTD:
	//case *Int:
	//case *Nat:
	//case *Raw:
	//case *Str:
	//case *Bnr:
	//case *Exc:
	//case *Msg:
	//case *Map:
	default:
		panic(errors.New(fmt.Sprintf("unhandeled type: %T", x)))
	}
	//return nil
}

func dotWalkType(b *dotBuilder, t Type) graph.Node {

	n, ok := b.typeCache[t]
	if ok {
		return n
	}
	switch x := t.(type) {
	case *EnumType:
		keys := make([]string, 0, len(*x))
		for k := range *x {
			keys = append(keys, k)
		}
		n := &dotNode{
			b.getNodeId(),
			keys,
			"EnumType",
		}
		b.nodes = append(b.nodes, n)
		for _, k := range keys {
			for _, inner_t := range (*x)[k] {
				v := dotWalkType(b, inner_t)
				e := &dotPortedEdge{
					id:       b.getEdgeId(),
					from:     n,
					to:       v,
					fromPort: k}
				b.edges = append(b.edges, e)
			}
		}
	case *IntType:
		n = &dotNode{
			b.getNodeId(),
			[]string{fmt.Sprintf("Size: %d", x.Size)},
			"IntType",
		}
		b.nodes = append(b.nodes, n)
	case *RawType:
		n = &dotNode{
			b.getNodeId(),
			[]string{fmt.Sprintf("Size: %d", x.Size)},
			"RawType",
		}
		b.nodes = append(b.nodes, n)
	case *NatType:
		n = &dotNode{
			b.getNodeId(),
			[]string{fmt.Sprintf("Size: %d", x.Size)},
			"NatType",
		}
		b.nodes = append(b.nodes, n)
	case *MapType:
		n = &dotNode{
			b.getNodeId(),
			[]string{"KeyType", "ValType"},
			"MapType",
		}
		b.nodes = append(b.nodes, n)
		kNode := dotWalkType(b, x.KeyType)
		ke := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       kNode,
			fromPort: "KeyType"}
		b.edges = append(b.edges, &ke)
		vNode := dotWalkType(b, x.ValType)
		ve := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       vNode,
			fromPort: "ValType"}
		b.edges = append(b.edges, &ve)
	case *AbsTT:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Vars", "Term"},
			"AbsTT",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Vars {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkType(b, &a),
				fromPort: "Vars"}
			b.edges = append(b.edges, &e)

		}
		e := &dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkType(b, x.Term),
			fromPort: "Term"}
		b.edges = append(b.edges, e)
	case *AppTT:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Args", "To"},
			"AppTT",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Args {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkType(b, a),
				fromPort: "Args"}
			b.edges = append(b.edges, &e)

		}
		e := &dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkType(b, x.To),
			fromPort: "To"}
		b.edges = append(b.edges, e)
	case *TypeVar:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Kind"},
			"TypeVar",
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkKind(b, x.Kind)
		e := &dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "Kind"}
		b.edges = append(b.edges, e)
	default:
		//return nil
		panic(errors.New(fmt.Sprintf("unhandeled type: %T", x)))
	}

	b.typeCache[t] = n
	return n
}

func dotWalkKind(b *dotBuilder, k Kind) graph.Node {
	n, ok := b.kindCache[k]
	if ok {
		return n
	}
	n = &dotNode{
		b.getNodeId(),
		[]string{},
		"Kind",
	}
	b.nodes = append(b.nodes, n)
	b.kindCache[k] = n
	return n
}

func dotWalkCond(b *dotBuilder, w *Cond) graph.Node {
	n := dotNode{
		b.getNodeId(),
		[]string{"Data", fmt.Sprintf("Case: %s", w.Case)},
		"Cond",
	}
	b.nodes = append(b.nodes, &n)
	for _, v := range w.Data {
		vNode := dotWalkBind(b, &v)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       vNode,
			fromPort: "Data"}
		b.edges = append(b.edges, &e)

	}
	return &n
}

func dotWalkBind(b *dotBuilder, d *Bind) graph.Node {
	n := dotNode{
		b.getNodeId(),
		[]string{"BindType", "Cond"},
		"Bind",
	}
	var m graph.Node
	m = dotWalkType(b, d.BindType)
	var e *dotPortedEdge
	e = &dotPortedEdge{
		id:       b.getEdgeId(),
		from:     n,
		to:       m,
		fromPort: "BindType"}
	b.edges = append(b.edges, e)
	if d.Cond != nil {
		m = dotWalkCond(b, d.Cond)
		e = &dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       m,
			fromPort: "Cond"}
		b.edges = append(b.edges, e)
	}
	return &n
}

func dotWalkDataCase(b *dotBuilder, d *DataCase) graph.Node {
	n := dotNode{
		b.getNodeId(),
		[]string{"Bind", "Body"},
		"DataCase",
	}
	var e *dotPortedEdge
	e = &dotPortedEdge{
		id:       b.getEdgeId(),
		from:     n,
		to:       dotWalkData(b, d.Body),
		fromPort: "Body"}
	b.edges = append(b.edges, e)
	e = &dotPortedEdge{
		id:       b.getEdgeId(),
		from:     n,
		to:       dotWalkBind(b, &d.Bind),
		fromPort: "Bind"}
	b.edges = append(b.edges, e)
	return &n
}

func dotWalkData(b *dotBuilder, d Data) graph.Node {
	n, ok := b.dataCache[d]
	if ok {
		return n
	}
	switch x := d.(type) {
	case *DataVar:
		n = &dotNode{
			b.getNodeId(),
			[]string{"DataType"},
			fmt.Sprintf("DataVar: %p", d),
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkType(b, x.DataType)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "DataType"}
		b.edges = append(b.edges, &e)
	case *PickData:
		n = &dotNode{
			b.getNodeId(),
			[]string{"From", "With"},
			"PickData",
		}
		b.nodes = append(b.nodes, n)
		for _, dc := range x.With {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkDataCase(b, &dc),
				fromPort: "With"}
			b.edges = append(b.edges, &e)

		}
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.From),
			fromPort: "From"}
		b.edges = append(b.edges, &e)
	case *Int:
		n = &dotNode{
			b.getNodeId(),
			[]string{"IntType", fmt.Sprintf("Data: %s", x.Data)},
			"Int",
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkType(b, x.IntType)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "IntType"}
		b.edges = append(b.edges, &e)
	case *Nat:
		n = &dotNode{
			b.getNodeId(),
			[]string{"NatType", fmt.Sprintf("Data: %s", x.Data)},
			"Nat",
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkType(b, x.NatType)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "NatType"}
		b.edges = append(b.edges, &e)
	case *Map:
		n = &dotNode{
			b.getNodeId(),
			[]string{"MapType", fmt.Sprintf("Data: %s", x.Data)},
			"Map",
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkType(b, x.MapType)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "MapType"}
		b.edges = append(b.edges, &e)
	case *AbsTD:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Vars", "Term"},
			"AbsTD",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Vars {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkType(b, &a),
				fromPort: "Vars"}
			b.edges = append(b.edges, &e)

		}
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.Term),
			fromPort: "Term"}
		b.edges = append(b.edges, &e)
	case *AbsDD:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Vars", "Term"},
			"AbsDD",
		}
		b.nodes = append(b.nodes, n)
		for _, v := range x.Vars {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkData(b, &v),
				fromPort: "Vars"}
			b.edges = append(b.edges, &e)

		}
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.Term),
			fromPort: "Term"}
		b.edges = append(b.edges, &e)
	case *AppTD:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Args", "To"},
			"AppTD",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Args {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkType(b, a),
				fromPort: "Args"}
			b.edges = append(b.edges, &e)

		}
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.To),
			fromPort: "To"}
		b.edges = append(b.edges, &e)
	case *AppDD:
		n = &dotNode{
			b.getNodeId(),
			[]string{"Args", "To"},
			"AppDD",
		}
		b.nodes = append(b.nodes, n)
		for _, a := range x.Args {
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       dotWalkData(b, a),
				fromPort: "Args"}
			b.edges = append(b.edges, &e)

		}
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       dotWalkData(b, x.To),
			fromPort: "To"}
		b.edges = append(b.edges, &e)
	case *Builtin:
		n = &dotNode{
			b.getNodeId(),
			[]string{"BuiltinType"},
			"Builtin",
		}
		b.nodes = append(b.nodes, n)
		tNode := dotWalkType(b, x.BuiltinType)
		e := dotPortedEdge{
			id:       b.getEdgeId(),
			from:     n,
			to:       tNode,
			fromPort: "BuiltinType"}
		b.edges = append(b.edges, &e)

	case *Proc:
		n = &dotNode{
			b.getNodeId(),
			[]string{},
			"Proc",
		}
		b.nodes = append(b.nodes, n)

		for i, _ := range x.Vars {
			v := &x.Vars[i]
			portName := fmt.Sprintf("Var %d", i)
			vNode := dotWalkData(b, v)
			n.ports = append(n.ports, portName)
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       vNode,
				fromPort: portName}
			b.edges = append(b.edges, &e)
		}

		for i, p := range x.Plan {
			portName := fmt.Sprintf("Unit %d", i)
			uNode := dotWalkUnit(b, p)
			n.ports = append(n.ports, portName)
			e := dotPortedEdge{
				id:       b.getEdgeId(),
				from:     n,
				to:       uNode,
				fromPort: portName}
			b.edges = append(b.edges, &e)
		}
	default:
		panic(errors.New(fmt.Sprintf("unhandeled type: %T", x)))
	}

	b.dataCache[d] = n
	return n
}

func (b *dotBuilder) getNodeId() int64 {
	id := b.nodeCounter
	b.nodeCounter++
	return id
}

func (b *dotBuilder) getEdgeId() int64 {
	id := b.edgeCounter
	b.edgeCounter++
	return id
}

func directedPortedAttrGraphFrom(b *dotBuilder) graph.Multigraph {
	dg := multi.NewDirectedGraph()
	for _, e := range b.edges {
		dg.SetLine(e)
	}
	return dg
}

func GetDot(b *CFGBuilder) string {
	//keys := make([]string, 0, len(b.GlobalVarMap))
	//for key := range b.GlobalVarMap {
	//keys = append(keys, key)
	//}
	d := dotBuilder{0, 0, []*dotNode{}, []*dotPortedEdge{}, map[Type]*dotNode{}, map[Data]*dotNode{}, map[Kind]*dotNode{}}
	//v, ok := stackMapPeek(b.fieldStack, "a")
	//if !ok {
	//panic(errors.New("var not found"))
	//}
	v := b.Constructor
	dotWalkData(&d, v)
	g := directedPortedAttrGraphFrom(&d)
	got, err := dot.MarshalMulti(g, "asd", "", "\t")
	_ = got
	if err != nil {
		panic(err)
	}
	return string(got)

}
