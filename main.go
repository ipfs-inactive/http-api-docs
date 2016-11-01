package main

import (
	"bytes"
	"fmt"
	"sort"

	cmds "github.com/ipfs/go-ipfs/commands"
	corecmds "github.com/ipfs/go-ipfs/core/commands"
)

const APIPrefix = "/api/v0"

var IgnoreEndpoints = map[string]bool{
	"/api/v0":        true,
	"/api/v0/files":  true,
	"/api/v0/pubsub": true,
}

type Endpoint struct {
	Name        string
	Arguments   []*Argument
	Options     []*Argument
	Description string
	Response    string
}

type Argument struct {
	Name        string
	Description string
	Type        string
	Required    bool
	Default     string
}

type Formatter interface {
	GenerateIntro() string
	GenerateIndex(endp []*Endpoint) string
	GenerateEndpointBlock(endp *Endpoint) string
	GenerateArgumentsBlock(args []*Argument, opts []*Argument) string
	GenerateBodyBlock(args []*Argument) string
	GenerateResponseBlock(response string) string
	GenerateExampleBlock(endp *Endpoint) string
}

func extractSubcommands(name string, cmd *cmds.Command) (endpoints []*Endpoint) {
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

	if ignore := IgnoreEndpoints[name]; !ignore {
		endpoints = []*Endpoint{
			&Endpoint{
				Name:        name,
				Description: cmd.Helptext.Tagline,
				Arguments:   arguments,
				Options:     options,
			},
		}
	}

	for n, cmd := range cmd.Subcommands {
		endpoints = append(endpoints,
			extractSubcommands(fmt.Sprintf("%s/%s", name, n), cmd)...)
	}
	return endpoints
}

func GenerateDocs(api []*Endpoint, formatter Formatter) string {
	buf := new(bytes.Buffer)
	buf.WriteString(formatter.GenerateIntro())
	buf.WriteString(formatter.GenerateIndex(api))
	for _, endp := range api {
		buf.WriteString(formatter.GenerateEndpointBlock(endp))
		buf.WriteString(formatter.GenerateArgumentsBlock(endp.Arguments, endp.Options))
		buf.WriteString(formatter.GenerateBodyBlock(endp.Arguments))
		buf.WriteString(formatter.GenerateResponseBlock(endp.Response))
		buf.WriteString(formatter.GenerateExampleBlock(endp))
	}
	return buf.String()
}

type byName []*Endpoint

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func main() {
	api := extractSubcommands(APIPrefix, corecmds.Root)
	sort.Sort(byName(api))

	formatter := new(MarkdownFormatter)
	fmt.Println(GenerateDocs(api, formatter))
}
