package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

type mockRenderer struct {
	report.Nodes
}

func (m mockRenderer) Render(rpt report.Report) report.Nodes { return m.Nodes }
func (m mockRenderer) Stats(rpt report.Report) render.Stats  { return render.Stats{} }

// RoughlyEqual compares the adjacencies and children of a sets of nodes,
// excluding their specific attributes. Useful when we want to check the rough
// shape of a set of nodes, but are not concerned with their specific
// properties.
func RoughlyEqual(a, b report.Nodes) bool {
	return reflect.DeepEqual(prune(a), prune(b))
}

// prune returns a copy of the Nodes with all information not strictly
// necessary for rendering nodes and edges in the UI cut away.
func prune(nodes report.Nodes) report.Nodes {
	result := report.Nodes{}
	for id, node := range nodes {
		result[id] = pruneNode(node)
	}
	return result
}

// pruneNode returns a copy of the Node with all information not strictly
// necessary for rendering nodes and edges stripped away. Specifically, that
// means cutting out parts of the Node.
func pruneNode(node report.Node) report.Node {
	prunedChildren := report.MakeNodeSet()
	node.Children.ForEach(func(child report.Node) {
		prunedChildren = prunedChildren.Add(pruneNode(child))
	})
	return report.MakeNode(
		node.ID).
		WithTopology(node.Topology).
		WithAdjacent(node.Adjacency.Copy()...).
		WithChildren(prunedChildren)
}

func TestReduceRender(t *testing.T) {
	renderer := render.Reduce([]render.Renderer{
		mockRenderer{Nodes: report.Nodes{"foo": report.MakeNode("foo")}},
		mockRenderer{Nodes: report.Nodes{"bar": report.MakeNode("bar")}},
	})

	want := report.Nodes{
		"foo": report.MakeNode("foo"),
		"bar": report.MakeNode("bar"),
	}
	have := renderer.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender1(t *testing.T) {
	// 1. Check when we return false, the node gets filtered out
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			return report.Nodes{}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
		}},
	}
	want := report.Nodes{}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestMapRender2(t *testing.T) {
	// 2. Check we can remap two nodes into one
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			return report.Nodes{
				"bar": report.MakeNode("bar"),
			}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo"),
			"baz": report.MakeNode("baz"),
		}},
	}
	want := report.Nodes{
		"bar": report.MakeNode("bar"),
	}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestMapRender3(t *testing.T) {
	// 3. Check we can remap adjacencies
	mapper := render.Map{
		MapFunc: func(nodes report.Node, _ report.Networks) report.Nodes {
			id := "_" + nodes.ID
			return report.Nodes{id: report.MakeNode(id)}
		},
		Renderer: mockRenderer{Nodes: report.Nodes{
			"foo": report.MakeNode("foo").WithAdjacent("baz"),
			"baz": report.MakeNode("baz").WithAdjacent("foo"),
		}},
	}
	want := report.Nodes{
		"_foo": report.MakeNode("_foo").WithAdjacent("_baz"),
		"_baz": report.MakeNode("_baz").WithAdjacent("_foo"),
	}
	have := mapper.Render(report.MakeReport())
	if !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func newu64(value uint64) *uint64 { return &value }
