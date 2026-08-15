package graph

// Builder accumulates nodes and edges, deduplicating nodes by ID so
// the same shared artifact (a favicon hash, a wallet address) is
// represented once even when many domains connect to it.
type Builder struct {
	g        Graph
	nodeSeen map[string]bool
	edgeSeen map[Edge]bool
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder {
	return &Builder{
		nodeSeen: make(map[string]bool),
		edgeSeen: make(map[Edge]bool),
	}
}

// AddNode adds a node with the given id, type, and label if id hasn't
// been added yet. Calling it again with the same id is a no-op — the
// first label wins — so callers don't need to check for existence
// themselves before wiring an edge to a shared node.
func (b *Builder) AddNode(id string, t NodeType, label string) {
	if b.nodeSeen[id] {
		return
	}
	b.nodeSeen[id] = true
	b.g.Nodes = append(b.g.Nodes, Node{ID: id, Type: t, Label: label})
}

// HasNode reports whether id has already been added.
func (b *Builder) HasNode(id string) bool {
	return b.nodeSeen[id]
}

// AddEdge adds a directed edge if the identical (from, to, type)
// triple hasn't been added yet.
func (b *Builder) AddEdge(from, to string, t EdgeType) {
	e := Edge{From: from, To: to, Type: t}
	if b.edgeSeen[e] {
		return
	}
	b.edgeSeen[e] = true
	b.g.Edges = append(b.g.Edges, e)
}

// Build returns the accumulated Graph.
func (b *Builder) Build() Graph {
	return b.g
}
