// Command genengine emits metric-specialized traversal loops.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type spec struct {
	Name, Kind, Query, Margin, Distance string
	Hamming                             bool
}

func main() {
	specs := []spec{
		{Name: "Angular", Kind: "Angular", Query: "q:=l.Floats(s.query,0); pp:=codec.F32(s.query,4)", Margin: "margin:=backend.Dot(l.Floats(node,0),q)", Distance: "v:=l.Floats(node,0); qq:=codec.F32(node,4); if qq==0 {qq=backend.Dot(v,v)}; raw:=angularRaw(backend.Dot(q,v),pp,qq)"},
		{Name: "Dot", Kind: "Dot", Query: "q:=l.Floats(s.query,0)", Margin: "margin:=backend.Dot(l.Floats(node,0),q)", Distance: "raw := -backend.Dot(q,l.Floats(node,0))"},
		{Name: "Euclidean", Kind: "Euclidean", Query: "q:=l.Floats(s.query,0)", Margin: "margin:=codec.F32(node,4)+backend.Dot(l.Floats(node,0),q)", Distance: "raw:=metric.Squared(q,l.Floats(node,0))"},
		{Name: "Manhattan", Kind: "Manhattan", Query: "q:=l.Floats(s.query,0)", Margin: "margin:=codec.F32(node,4)+backend.Dot(l.Floats(node,0),q)", Distance: "raw:=metric.L1(q,l.Floats(node,0))"},
		{Name: "Hamming", Kind: "Hamming", Hamming: true, Query: "q:=l.WordsAt(s.query,0)", Margin: "bit:=l.WordsAt(node,0)[0]; margin:=int((q[bit/64]>>(63-bit%64))&1)", Distance: "v:=l.WordsAt(node,0); distance:=0;for j,x:=range q{distance+=bits.OnesCount64(x^v[j])};raw:=float32(distance)"},
	}
	t := template.Must(template.New("search").Parse(searchTemplate))
	for _, s := range specs {
		var b bytes.Buffer
		if err := t.Execute(&b, s); err != nil {
			panic(err)
		}
		code, err := format.Source(b.Bytes())
		if err != nil {
			panic(fmt.Sprintf("%s: %v", s.Name, err))
		}
		if err := os.WriteFile("search_"+strings.ToLower(s.Name)+".go", code, 0644); err != nil {
			panic(err)
		}
	}
}
