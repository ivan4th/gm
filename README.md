# Golden Tests for Go

This is `gm` Golden test framework extracted from
[Virtlet](https://github.com/Mirantis/virtlet/)

The basic idea of this library is that in any Go test you can use
`Verify()` function to compare some data against existing golden
copy stored in Git index (that is, staged or a part of the
most recent commit's tree). If the golden copy doesn't exist or differs,
the test fails, and the working copy of the file will contain the new
data. If you judge that the updated file is correct and then stage or
commit the updated file, the test will pass.

**WARNING** Golden tests cannot replace normal tests under all
circumstances. Diffs are error-prone, especially when they're large,
you can easily miss important problems when looking at them. Also,
beware of non-deterministic output, such as caused by using current
system time, RNG (e.g. to generate UUIDs), or having your output
depend on the order of map traversal. Making such code compatible with
the golden tests requires some extra effort (fake clock, fake random,
sorting map keys, etc.).

An sample test is below (see
[examples/my_test.go](examples/my_test.go)).  If you run it right from
`examples/` directory, you might want to remove `*.out.*` files first
try out things from scratch. The tests don't have any logic to
generate the data being emitted, so we just dump some string or struct
literals, but in the real-world cases, there will be some non-trivial
code under test that produces these data.

```go
package mypackage

import (
	"github.com/ivan4th/gm"
	"testing"
)

func TestSomething(t *testing.T) {
	myText := "abc\ndef"
	gm.Verify(t, myText)
}
```

Try running the test:

```console
$ go test
--- FAIL: TestSomething (0.01s)
    gm.go:72: got difference for "TestSomething" ("/Users/ivan4th/work/gm/examples/TestSomething.out.txt"):
        <NEW FILE>
        abc
        def
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.124s
```

Oops, there are some changes, namely, a new golden file is created,
let's accept it as a correct one:

```console
$ git add TestSomething.out.txt
$ go test
PASS
ok      github.com/ivan4th/gm/examples  0.116s
```

Then, let's "break" it again by changing the output in the code to see
what happens:

```console
$ sed -i s/abc/qqq/ my_test.go
$ go test -run '^TestSomething$'
--- FAIL: TestSomething (0.01s)
    gm.go:72: got difference for "TestSomething" ("/Users/ivan4th/work/gm/examples/TestSomething.out.txt"):
        diff --git a/examples/TestSomething.out.txt b/examples/TestSomething.out.txt
        index 85137a6..8fca6eb 100755
        --- a/examples/TestSomething.out.txt
        +++ b/examples/TestSomething.out.txt
        @@ -1,2 +1,2 @@
        -abc
        +qqq
         def
        \ No newline at end of file
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.125s
```

And now let's fix it:
```console
$ git add TestSomething.out.txt
$ go test -run '^TestSomething$'
PASS
ok      github.com/ivan4th/gm/examples  0.114s
```

Besides emitting simple text, you can also emit JSON or YAML
serializations of the objects. There's also a possibility to emit
custom serializations using `gm.Verifier` interface. Objects that are
not `string`, `[]byte` and do not implement `gm.Verifier` interface
are serialized as JSON:

```go
type Foo struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func TestJSON(t *testing.T) {
	gm.Verify(t, Foo{Name: "rrr", ID: 42})
}
```

Now let's try it:

```console
$ go test -run '^TestJSON$'
--- FAIL: TestJSON (0.01s)
    gm.go:72: got difference for "TestJSON" ("/Users/ivan4th/work/gm/examples/TestJSON.out.json"):
        <NEW FILE>
        {
          "name": "rrr",
          "id": 42
        }
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.116s
$ git add TestJSON.out.json
$ go test -run '^TestJSON$'
PASS
ok      github.com/ivan4th/gm/examples  0.123s
$ sed -i s/rrr/fff/ my_test.go
$ go test -run '^TestJSON$'
--- FAIL: TestJSON (0.01s)
    gm.go:72: got difference for "TestJSON" ("/Users/ivan4th/work/gm/examples/TestJSON.out.json"):
        diff --git a/examples/TestJSON.out.json b/examples/TestJSON.out.json
        index ae7a4c6..0cb4514 100755
        --- a/examples/TestJSON.out.json
        +++ b/examples/TestJSON.out.json
        @@ -1,4 +1,4 @@
         {
        -  "name": "rrr",
        +  "name": "fff",
           "id": 42
         }
        \ No newline at end of file
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.119s
$ git add TestJSON.out.json
$ go test -run '^TestJSON$'
PASS
ok      github.com/ivan4th/gm/examples  0.135s
```

It is also possible to emit YAML output. `gm` uses
[github.com/ghodss/yaml](https://github.com/ghodss/yaml) library for
YAML serialization, which uses intermediate JSON representation and
thus is compatible with `json` struct tags, which are used, for
example, in Kubernetes API objects (such as Custom Resources):

```go
func TestYAML(t *testing.T) {
	foo := Foo{Name: "rrr", ID: 42}
	gm.Verify(t, gm.NewYamlVerifier(foo))
}
```

Let's try it:

```console
$ go test -run '^TestYAML$'
--- FAIL: TestYAML (0.01s)
    gm.go:72: got difference for "TestYAML" ("/Users/ivan4th/work/gm/examples/TestYAML.out.yaml"):
        <NEW FILE>
        id: 42
        name: rrr
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.121s
$ git add TestYAML.out.yaml
$ go test -run '^TestYAML$'
PASS
ok      github.com/ivan4th/gm/examples  0.125s
$ sed -i s/rrr/fff/ my_test.go
$ go test -run '^TestYAML$'
--- FAIL: TestYAML (0.01s)
    gm.go:72: got difference for "TestYAML" ("/Users/ivan4th/work/gm/examples/TestYAML.out.yaml"):
        diff --git a/examples/TestYAML.out.yaml b/examples/TestYAML.out.yaml
        index 0c8206d..9ea7f5f 100755
        --- a/examples/TestYAML.out.yaml
        +++ b/examples/TestYAML.out.yaml
        @@ -1,2 +1,2 @@
         id: 42
        -name: rrr
        +name: fff
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.124s
```

In some cases, you might want to replace a string in the output using
text substitution. This might be helpful if, for example, your output
contains some temporary directory names, etc. A simple substitution
example which replaces the string `rrr` with `qqq`:

```go
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
```

If we run it, we'll see that the output contains "bbb" in place of
"rrr" which is serialized from the struct field:

```console
$ go test -run '^TestSubst$'
--- FAIL: TestSubst (0.01s)
    gm.go:72: got difference for "TestSubst" ("/Users/ivan4th/work/gm/examples/TestSubst.out.yaml"):
        <NEW FILE>
        id: 42
        name: bbb
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.123s
```

`gm` is compatible with table-driven tests:

```go
$ go test -run '^TestTable$'
--- FAIL: TestTable (0.03s)
    --- FAIL: TestTable/case_one (0.02s)
        gm.go:72: got difference for "TestTable/case_one" ("/Users/ivan4th/work/gm/examples/TestTable__case_one.out.yaml"):
            <NEW FILE>
            id: 42
            name: aaa
    --- FAIL: TestTable/case_two (0.01s)
        gm.go:72: got difference for "TestTable/case_two" ("/Users/ivan4th/work/gm/examples/TestTable__case_two.out.yaml"):
            <NEW FILE>
            id: 4242
            name: ccc
FAIL
exit status 1
FAIL    github.com/ivan4th/gm/examples  0.148s
$ git add .
$ go test -run '^TestTable$'
PASS
ok      github.com/ivan4th/gm/examples  0.135s
```
