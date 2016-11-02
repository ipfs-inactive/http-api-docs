// Package docs can be used to gather go-ipfs commands and automatically
// generate documentation or tests.
package docs

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"

	cmds "github.com/ipfs/go-ipfs/commands"
	corecmds "github.com/ipfs/go-ipfs/core/commands"
)

// A map of single endpoints to be skipped (subcommands are processed though).
var IgnoreEndpoints = map[string]bool{}

// Endpoint defines an IPFS RPC API endpoint.
type Endpoint struct {
	Name        string
	Arguments   []*Argument
	Options     []*Argument
	Description string
	Response    *Response
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

type Response struct {
	Text   bool
	Schema string
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

// Endpoints receives a name and a go-ipfs command and returns the endpoints it
// defines] (sorted). It does this by recursively gathering endpoints defined by
// subcommands. Thus, calling it with the core command Root generates all
// the endpoints.
func Endpoints(name string, cmd *cmds.Command) (endpoints []*Endpoint) {
	var arguments []*Argument
	for _, arg := range cmd.Arguments {
		argType := "string"
		if arg.Type == cmds.ArgFile {
			argType = "file"
		}
		arguments = append(arguments, &Argument{
			Name:        arg.Name,
			Type:        argType,
			Required:    arg.Required,
			Description: arg.Description,
		})
	}

	var options []*Argument
	for _, opt := range cmd.Options {
		def := fmt.Sprint(opt.DefaultVal())
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

	ignore := len(cmd.Subcommands) > 0 || IgnoreEndpoints[name]
	if !ignore {
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

func interfaceToJsonish(t reflect.Type, i int) string {
	// Aux function
	insertIndent := func(i int) string {
		buf := new(bytes.Buffer)
		for j := 0; j < i; j++ {
			buf.WriteRune(' ')
		}
		return buf.String()
	}

	result := new(bytes.Buffer)
	if i > 20 { // 5 levels is enough. Infinite loop failsafe
		return insertIndent(i) + "...\n"
	}

	switch t.Kind() {
	case reflect.Invalid:
		result.WriteString("null\n")
	case reflect.Ptr:
		return interfaceToJsonish(t.Elem(), i)
	case reflect.Map:
		result.WriteString(insertIndent(i) + "{\n")
		result.WriteString(insertIndent(i+4) + fmt.Sprintf(`"<%s>": `, t.Key().Kind()))
		result.WriteString(interfaceToJsonish(t.Elem(), i+4))
		result.WriteString(insertIndent(i) + "}\n")
	case reflect.Struct:
		result.WriteString(insertIndent(i) + "{\n")
		for j := 0; j < t.NumField(); j++ {
			f := t.Field(j)
			result.WriteString(fmt.Sprintf(insertIndent(i+4)+"\"%s\": ", f.Name))
			result.WriteString(interfaceToJsonish(f.Type, i+4))
		}
		result.WriteString(insertIndent(i) + "}\n")
	case reflect.Slice:
		sType := t.Elem().Kind().String()
		if sType == "ptr" {
			result.WriteString("null\n")
		} else {
			result.WriteString("[\n")
			result.WriteString(interfaceToJsonish(t.Elem(), i+4))
			result.WriteString(insertIndent(i) + "]\n")
		}
	default:
		result.WriteString(insertIndent(i) + "\"<" + t.Kind().String() + ">\"\n")

	}

	fix, _ := regexp.Compile(":[ ]+")
	finalResult := string(fix.ReplaceAll(result.Bytes(), []byte(": ")))

	return string(finalResult)
}

func buildResponse(res interface{}) *Response {
	// Commands with a nil type return text. This is a bad thing.
	if res == nil {
		return &Response{
			Text: true,
		}
	}

	return &Response{
		Text:   false,
		Schema: interfaceToJsonish(reflect.TypeOf(res), 0) + "\n",
	}
}
