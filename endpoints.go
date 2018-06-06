// Package docs can be used to gather go-ipfs commands and automatically
// generate documentation or tests.
package docs

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	cmds "gx/ipfs/QmTjNRVt2fvaRFu93keEC7z5M1GS1iH6qZ9227htQioTUY/go-ipfs-cmds"
	corecmds "gx/ipfs/QmcKwjeebv5SX3VFUGDFa4BNMYhy14RRaCzQP7JN3UQDpB/go-ipfs/core/commands"
	config "gx/ipfs/QmcKwjeebv5SX3VFUGDFa4BNMYhy14RRaCzQP7JN3UQDpB/go-ipfs/repo/config"
	cmdkit "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"
)

// A map of single endpoints to be skipped (subcommands are processed though).
var IgnoreEndpoints = map[string]bool{}

// How much to indent when generating the response schemas
const IndentLevel = 4

// Failsafe when traversing objects containing objects of the same type
const MaxIndent = 20

// Endpoint defines an IPFS RPC API endpoint.
type Endpoint struct {
	Name        string
	Arguments   []*Argument
	Options     []*Argument
	Description string
	Response    string
	Group       string
}

// Argument defines an IPFS RPC API endpoint argument.
type Argument struct {
	Name        string
	Description string
	Type        string
	Required    bool
	Default     string
}

type sorter []*Endpoint

func (a sorter) Len() int           { return len(a) }
func (a sorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

const APIPrefix = "/api/v0"

// AllEndpoints gathers all the endpoints from go-ipfs.
func AllEndpoints() []*Endpoint {
	return Endpoints(APIPrefix, corecmds.Root)
}

func IPFSVersion() string {
	return config.CurrentVersionNumber
}

// Endpoints receives a name and a go-ipfs command and returns the endpoints it
// defines] (sorted). It does this by recursively gathering endpoints defined by
// subcommands. Thus, calling it with the core command Root generates all
// the endpoints.
func Endpoints(name string, cmd *cmds.Command) (endpoints []*Endpoint) {
	var arguments []*Argument
	var options []*Argument

	ignore := len(cmd.Subcommands) > 0 || IgnoreEndpoints[name]
	if !ignore { // Extract arguments, options...
		for _, arg := range cmd.Arguments {
			argType := "string"
			if arg.Type == cmdkit.ArgFile {
				argType = "file"
			}
			arguments = append(arguments, &Argument{
				Name:        arg.Name,
				Type:        argType,
				Required:    arg.Required,
				Description: arg.Description,
			})
		}

		for _, opt := range cmd.Options {
			def := fmt.Sprint(opt.Default())
			if def == "<nil>" {
				def = ""
			}
			options = append(options, &Argument{
				Name:        opt.Names()[0],
				Type:        opt.Type().String(),
				Description: opt.Description(),
				Default:     def,
			})
		}

		res := buildResponse(cmd.Type)

		endpoints = []*Endpoint{
			&Endpoint{
				Name:        name,
				Description: cmd.Helptext.Tagline,
				Arguments:   arguments,
				Options:     options,
				Response:    res,
			},
		}
	}

	for n, cmd := range cmd.Subcommands {
		endpoints = append(endpoints,
			Endpoints(fmt.Sprintf("%s/%s", name, n), cmd)...)
	}
	sort.Sort(sorter(endpoints))
	return endpoints
}

// TODO: This maybe should be a separate Go module for reusability
func interfaceToJsonish(t reflect.Type, i int) string {
	// Aux function
	insertIndent := func(i int) string {
		buf := new(bytes.Buffer)
		for j := 0; j < i; j++ {
			buf.WriteRune(' ')
		}
		return buf.String()
	}

	countExported := func(t reflect.Type) int {
		if t.Kind() != reflect.Struct {
			return 0
		}

		count := 0

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Name[0:1] == strings.ToUpper(f.Name[0:1]) {
				count++
			}
		}
		return count
	}

	result := new(bytes.Buffer)
	if i > MaxIndent { // 5 levels is enough. Infinite loop failsafe
		return insertIndent(i) + "...\n"
	}

	switch t.Kind() {
	case reflect.Invalid:
		result.WriteString("null\n")
	case reflect.Interface:
		description := "\"<object>\""
		// Handle types that should be called out specially
		if t.String() == "multiaddr.Multiaddr" {
			description = "\"<multiaddr-string>\""
		}
		result.WriteString(fmt.Sprintf("%s%s\n", insertIndent(i), description))
	case reflect.Ptr:
		// CIDs have specialized JSON marshaling, see:
		// https://github.com/ipfs/go-cid/blob/078355866b1dda1658b5fdc5496ed7e25fdcf883/cid.go#L407-L415
		if t.String() == "*cid.Cid" {
			result.WriteString(insertIndent(i) + "{ \"/\": \"<cid-string>\" }\n")
		} else if _, ok := t.MethodByName("String"); ok && countExported(t.Elem()) == 0 {
			return interfaceToJsonish(reflect.TypeOf(""), i)
		} else {
			return interfaceToJsonish(t.Elem(), i)
		}
	case reflect.Map:
		result.WriteString(insertIndent(i) + "{\n")
		result.WriteString(insertIndent(i+IndentLevel) + fmt.Sprintf(`"<%s>": `, t.Key().Kind()))
		result.WriteString(interfaceToJsonish(t.Elem(), i+IndentLevel))
		result.WriteString(insertIndent(i) + "}\n")
	case reflect.Struct:
		if _, ok := t.MethodByName("String"); ok && countExported(t) == 0 {
			return interfaceToJsonish(reflect.TypeOf(""), i)
		}
		result.WriteString(insertIndent(i) + "{\n")
		for j := 0; j < t.NumField(); j++ {
			f := t.Field(j)
			result.WriteString(fmt.Sprintf(insertIndent(i+IndentLevel)+"\"%s\": ", f.Name))
			result.WriteString(interfaceToJsonish(f.Type, i+IndentLevel))
		}
		result.WriteString(insertIndent(i) + "}\n")
	case reflect.Slice:
		result.WriteString("[\n")
		result.WriteString(interfaceToJsonish(t.Elem(), i+IndentLevel))
		result.WriteString(insertIndent(i) + "]\n")
	default:
		result.WriteString(insertIndent(i) + "\"<" + t.Kind().String() + ">\"\n")

	}

	// This removes wrong indents in cases like "key:      <string>"
	fix, _ := regexp.Compile(":[ ]+")
	finalResult := string(fix.ReplaceAll(result.Bytes(), []byte(": ")))

	return string(finalResult)
}

func buildResponse(res interface{}) string {
	// Commands with a nil type return text. This is a bad thing.
	if res == nil {
		return "This endpoint returns a `text/plain` response body."
	}

	return interfaceToJsonish(reflect.TypeOf(res), 0)
}
