package examples

import (
	"github.com/ivan4th/gm"
	"testing"
)

func TestSomething(t *testing.T) {
	myText := "abc\ndef"
	gm.Verify(t, myText)
}

type Foo struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func TestJSON(t *testing.T) {
	gm.Verify(t, Foo{Name: "rrr", ID: 42})
}

func TestYAML(t *testing.T) {
	foo := Foo{Name: "rrr", ID: 42}
	gm.Verify(t, gm.NewYamlVerifier(foo))
}

func TestSubst(t *testing.T) {
	foo := Foo{Name: "rrr", ID: 42}
	gm.Verify(t, gm.NewSubstVerifier(
		gm.NewYamlVerifier(foo),
		[]gm.Replacement{
			{
				Old: "rrr",
				New: "bbb",
			},
		}))
}

func TestTable(t *testing.T) {
	for _, tc := range []struct {
		name string
		foo  Foo
	}{
		{
			name: "case one",
			foo:  Foo{Name: "aaa", ID: 42},
		},
		{
			name: "case two",
			foo:  Foo{Name: "ccc", ID: 4242},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gm.Verify(t, gm.NewYamlVerifier(tc.foo))
		})
	}
}
