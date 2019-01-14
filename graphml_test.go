package graphml

import (
	"compress/gzip"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testdata = "data"

func TestRoundtrip(t *testing.T) {
	const ext = Ext + ".gz"

	dir, err := os.Open(testdata)
	require.NoError(t, err)
	defer dir.Close()

	for {
		names, err := dir.Readdirnames(100)
		if err == io.EOF {
			err = nil
		}
		require.NoError(t, err)
		if len(names) == 0 {
			break
		}
		for _, name := range names {
			if !strings.HasSuffix(name, ext) {
				continue
			}
			name := name
			t.Run(strings.TrimSuffix(name, ext), func(t *testing.T) {
				name = filepath.Join(testdata, name)
				f, err := os.Open(name)
				require.NoError(t, err)
				defer f.Close()

				zr, err := gzip.NewReader(f)
				require.NoError(t, err)
				defer zr.Close()

				doc, err := Decode(zr)
				require.NoError(t, err)

				out, err := os.Create(strings.TrimSuffix(name, ".gz"))
				require.NoError(t, err)
				defer out.Close()

				err = Encode(out, doc)
				require.NoError(t, err)
			})
		}
	}
}

func createEdge(i, s, t int) Edge {
	e := NewEdge(fmt.Sprintf("e%02d", i), fmt.Sprintf("n%02d", s), fmt.Sprintf("n%02d", t))
	e.Data = []Data{NewData("w", float32(i+1)/4)}
	return e
}

func TestManualDocumentCreation(t *testing.T) {
	doc := new(Document)
	doc.Instr.Target = "xml"
	doc.Instr.Inst = []byte("version=\"1.0\" encoding=\"UTF-8\"")
	doc.Keys = []Key{
		NewKey(KindNode, "n", "label", "string"),
		NewKey(KindNode, "c", "cute", "boolean"),
		NewKey(KindNode, "s", "size", "int"),
		NewKey(KindEdge, "w", "weight", "float")}
	g := Graph{}
	g.EdgeDefault = EdgeDirected
	names := []string{"Gizmo", "Gopher", "Gong", "Gonzo", "Gracie", "Granite", "Gobi"}
	g.Nodes = make([]Node, len(names))
	for i, name := range names {
		n := Node{}
		n.ID = fmt.Sprintf("n%02d", i)
		n.Data = []Data{
			NewData("n", name),
			NewData("c", true),
			NewData("s", (10-i)*3),
		}
		g.Nodes[i] = n
	}
	g.Edges = []Edge{
		createEdge(0, 0, 1),
		createEdge(1, 0, 4),
		createEdge(2, 2, 3),
		createEdge(3, 2, 5),
	}
	doc.Graphs = []Graph{g}

	out, err := os.Create("data/GNames.graphml")
	require.NoError(t, err)
	defer out.Close()
	err = Encode(out, doc)
	require.NoError(t, err)

}
