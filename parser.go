package main

import (
	"os"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

// A valid indentifier part is one of the following:
// 1. an escaped character, like \"
// 2. a string of characters
var re_valid_ident_part = `(\\.|[a-zA-Z_][-a-zA-Z\d_]+)`

// GoFigureLexer - Contains the lexicographic rules for how gofigure is parsed
var GoFigureLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|([#;].*$)|(/\*[.\s\n\r]*\*/)` + // Comments
		`|(?P<MLString>("""(?:\\.|[^(""")])*""")|('''(?:\\.|[^(''')])*'''))` +
		`|(?P<String>("(\\"|[^"])*?")|('(\\'|[^'])*?'))` +
		`|(?P<Boolean>true|false)` +
		`|(?P<Ident>` + re_valid_ident_part + `)` +
		`|(?P<Float>-?\d+\.\d+)` +
		`|(?P<Int>-?\d+)` +
		`|(?P<SectionEnd>\[\])` +
		`|(?P<Include>%include)` +
		`|(?P<Expand>\.\.\.)` +
		`|(?P<Special>[][{}.,:%@])`,
))

// FigureConfig - Structure capable of containing a full GoFigure configuration
type FigureConfig struct {
	Entries []*Entry `(@@)*`

	Pos lexer.Position
}

type Entry struct {
	Include *Include `@@`
	Section *Section `| "[" @@ (SectionEnd|EOF)?`
	Field   *Field   `| @@`
	Pos     lexer.Position
}

type Include struct {
	Includes []string `"%include" @String (","? @String)* `

	Pos lexer.Position
}

/*
	As SectionRoot and SectionChild must have different rules for how they are parsed,
	they have to be separate structres.
*/

type Section struct {
	Roots  []SectionRoot `(@@)+ "]"`
	Fields []*Field      `(@@)*`

	Pos lexer.Position
}

type SectionRoot struct {
	Identifier []string      `(@(Ident|"@") ("," " "*|" ")? | "%" "{" (@Ident ("," " "*|" ")?)* "}")`
	Child      *SectionChild `(@@)?`

	Pos lexer.Position
}

type SectionChild struct {
	Identifier []string      `"." (@(Ident|"@"|Int) ("," " "*|" ")? | "%" "{" (@Int @"..." @Int | (@(Ident|Int) ("," " "*|" ")?)*) "}")`
	Child      *SectionChild `(@@)?`

	Pos lexer.Position
}

type Field struct {
	Key   string      `( (@Ident|@String) ` // Key
	Child *ChildField `	( "." @@`           // When a child field should be created this is where it goes
	Value *Value      `	| ":" @@ )?)`       // ? == allow empty values

	ArrayIndex *int64
	// ArrayIndex is not populated at parse-time,
	// it's in this struct as childfields later get expanded to regular fields

	Pos lexer.Position
}

type ChildField struct {
	Key        string      `(( (@Ident|@String) ` // Key
	ArrayIndex *int64      `|@Int)`
	Child      *ChildField `( "." @@`     // When a child field should be created this is where it goes
	Value      *Value      `| ":" @@ )?)` // ? == allow empty values

	Pos lexer.Position
}

type UnprocessedString struct {
	String *string `@MLString`

	Pos lexer.Position
}

type Bool bool

func (b *Bool) Capture(v []string) error { *b = v[0] == "true"; return nil }

type Value struct {
	String          *string            `@String`
	MultilineString *UnprocessedString `| @@`
	Integer         *int64             `| @Int`
	Float           *float64           `| @Float`
	Boolean         *Bool              `| (@"true" | @"false") `
	Map             []*Field           `| "{" ((@@ ","?)* )? "}"`
	ParsedArray     []*Value           `| "[" ((@@ ","?)* )? "]"`
	Identifier      *string            `| @Ident @("." Ident)*`

	// Here is where a sequential-number named map goes
	FinalArray []*Value

	Pos lexer.Position
}

func checkFileError(err error, filename string) {
	if err != nil {
		panic(strings.Replace(err.Error(), "<source>", filename, 1))
	}
}

var _workingDirectory string

func setWorkingDirectory(workingDirectory string) {
	_workingDirectory = workingDirectory
}

// BuildParser - Builds a new parser with GoFigureLexer as lexer
func BuildParser() (parser *participle.Parser) {
	parser, err := participle.Build(
		&FigureConfig{},
		participle.Lexer(GoFigureLexer),
		participle.Unquote("String"),
	)

	check(err)

	return
}

var fileCache map[string]FigureConfig

// ParseFile - Parses a file with given filename and parser. If a nil argument is passed instead of a parser a new one is built
func ParseFile(filename string, parser *participle.Parser) (config FigureConfig) {
	if config, exists := fileCache[filename]; exists {
		return config
	}

	config = FigureConfig{}
	if parser == nil {
		parser = BuildParser()
	}

	if _workingDirectory != "" {
		filename = _workingDirectory + "/" + filename
	}

	// Open a handle to file
	file, err := os.Open(filename)

	check(err)

	err = parser.Parse(file, &config)

	check(file.Close())

	checkFileError(err, filename)

	return
}
