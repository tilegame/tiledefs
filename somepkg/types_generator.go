// +build generator

package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

// TileGen is a structure used to automatically generate the TileType
// definitions and the GameWorldGrid arrays.  This is used by the
// template to produce Go source code.
type TileGen struct {

	// Timestamp for when the generation was done.
	Timestamp string

	// List of tile definitions extracted from the text source.
	Definitions []TileDef

	// PropertyNameSet is a unique list of property names.
	PropertyNameSet map[string]bool

	// DefaultProperties helps map tile types to their properties.
	// It's a string containing go code.
	DefaultProperties string
}

type TileDef struct {
	Symbol     string
	Name       string
	Properties []string
}

// ReadmeTemplate is the template for the generated README.md file.
const ReadmeTemplate = `
Tile Types Definitions
----------------------------------------------------------------------
 
Time Generated |
---------------|
{{.Timestamp}} |

This defintion table is generated from tile_types.txt.  Each symbol
corresponds to the Name, which becomes a constant in types.go.  Since
the following symbols have been defined, they can be used in grid.txt
to create game worlds!
 
| Symbol | Name | Default Properties  |
|:------:|------|---------------------|
{{range $tile := .Definitions}}| {{$tile.Symbol}} | {{$tile.Name}} | {{range $p := $tile.Properties}} {{$p}},{{end}} |
{{end}}

`

// CodeTemplate is the template for the automatically generated go
// source code.
const CodeTemplate = `
// Code generated by types_generator.go - DO NOT EDIT.

package somepkg

const (
	_ Kind = iota
	{{range $tile := .Definitions}}{{$tile.Name}}
	{{end}}
)

const (
	_ Property = (1 << iota) >> 1
	{{range $name, $_ := .PropertyNameSet}}{{$name}}
	{{end}}
)

var DefaultProperty = map[Kind]Property {
{{.DefaultProperties}}
}

var SymbolToKind = map[string]Kind {
	{{range $tile := .Definitions}}"{{$tile.Symbol}}": {{$tile.Name}},
	{{end}}
}
`

const (
	filenameTileTypes  = "src/types.txt"
	filenameGrid       = "src/grid.txt"
	filenameCodeOutput = "types.go"
	filenameReadme     = "readme.md"
)

var (
	tilegen = &TileGen{
		PropertyNameSet: make(map[string]bool),
	}
)

// reads the file into a string, then calls the parser
func processTileTypesFile() {
	b, err := ioutil.ReadFile(filenameTileTypes)
	if err != nil {
		log.Fatal(err)
	}
	parseTileDefs(string(b))
}

// The parser that scans line-by-line, converting each valid line into
func parseTileDefs(s string) {

	scanner := bufio.NewScanner(strings.NewReader(s))
	line := ""

	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		arr := strings.Split(line, " ")

		if len(arr) < 2 {
			continue
		}

		symbol := arr[0]
		name := capitilizeFirstLetter(arr[1])
		proplist := []string{}

		// If there aren't any properties, finish up.
		if len(arr) == 2 {
			goto finish
		}

		// If there are, capitilize their names,
		// and add them to the set of all properties.
		for _, v := range arr[2:] {
			prop := capitilizeFirstLetter(v)
			proplist = append(proplist, prop)
			tilegen.PropertyNameSet[prop] = true
		}

	finish:
		// Add this tile definition to the list.
		tile := TileDef{
			Symbol:     symbol,
			Name:       name,
			Properties: proplist,
		}
		tilegen.Definitions = append(tilegen.Definitions, tile)
	}
}

// capitilizing the first letter is needed in order to make the
// TileType an exported value.
func capitilizeFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	first := strings.ToUpper(s[0:1])
	return first + s[1:]
}

// stringifyPropertyList converts the propertymap into a string of go
// code.  If no properties are listed, then it defaults to 0.  If
// there are, then the properties are combined into a bitmask using
// Alternation.
//
// In other words, for each property, the OR operator is appended to
// each property name in the list, except for the very last one.
//
func (tg *TileGen) stringifyPropertyList() {
	s := ""
	for _, tile := range tg.Definitions {

		list := tile.Properties
		s += "\t" + tile.Name + ": "

		switch len(list) {
		case 0:
			s += "0,\n"
			continue

		case 1:
			s += list[0] + ",\n"
			continue
		}

		// For arrays greater than length of 1, we will split
		// off the final element.  the | symbol goes after
		// each element except for the last one.
		arr, last := list[:len(list)-1], list[len(list)-1]

		for _, v := range arr {
			s += v + " | "
		}

		s += last + ",\n"
	}
	tg.DefaultProperties = s
}

func makeTheFile(templ, filename string) {
	t := template.Must(template.New(filename).Parse(templ))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	os.Truncate(filename, 0)
	t.Execute(file, tilegen)
}

func main() {
	tilegen.Timestamp = time.Now().Format(time.UnixDate)
	processTileTypesFile()
	tilegen.stringifyPropertyList()
	makeTheFile(ReadmeTemplate, filenameReadme)
	makeTheFile(CodeTemplate, filenameCodeOutput)
}
