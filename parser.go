package main

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var PutinLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(?P<Include>#include)` +
		`|([#;].*$)` +
		`|(?P<MLString>("""(?:\\.|[^(""")])*""")|('''(?:\\.|[^(''')])*'''))` +
		`|(?P<String>("(?:\\.|[^"])*")|('(?:\\.|[^'])*'))` +
		`|(?P<Ident>(\\.|(@|[[:alpha:]_][[:alpha:]\d_]+)(\.(@|[[:alpha:]\d_]+))*\.?))` +
		`|(?P<Float>-?\d+\.\d+)` +
		`|(?P<Int>-?\d+)` +
		`|(?P<Punct>[][{},. :%])` +
		`|(?P<FileEnd>\z)`,
))

type CONFIG struct {
	Entries []*Entry `(@@)*`

	Pos lexer.Position
}

type Entry struct {
	Include *Include `@@`
	Field   *Field   `| @@`
	Section *Section `| @@`

	Pos lexer.Position
}

type Include struct {
	Includes []string `"#include" @String (","? @String)*`
	Parsed   []*CONFIG

	Pos lexer.Position
}

type Section struct {
	Identifier []string `"[" (@(Ident " ")+ | (@(Ident? ("%" "{" (","? Ident)* "}")? "."? Ident? ("%" "{" (","? Ident)* "}")?)!)* ) "]"`
	Fields     []*Field ` (@@)* ("[]"|FileEnd)?`

	Pos lexer.Position
}

type Field struct {
	Key   string `@Ident (":"`
	Value *Value `@@)?`

	Pos lexer.Position
}

type UnprocessedString struct {
	String *string `@MLString`
}

type Value struct {
	String          *string            `@String`
	MultilineString *UnprocessedString `| @@`
	Integer         *int64             `| @Int`
	Float           *float64           `| @Float`
	Identifier      *string            `| @Ident `
	List            []*Value           `| "[" ((@@)*)? "]"`
	Map             []*Field           `| "{" ((@@)*)? "}"`

	Reassigns bool

	Pos lexer.Position
}

func checkFileError(err error, filename string) {
	if err != nil {
		panic(strings.Replace(err.Error(), "<source>", filename, -1))
	}
}

func BuildParser() (parser *participle.Parser) {
	parser, err := participle.Build(
		&CONFIG{},
		participle.Lexer(PutinLexer),
		participle.Unquote("String"),
	)

	check(err)

	return
}

func ParseFile(filename string, parser *participle.Parser) (config *CONFIG) {
	config = &CONFIG{}
	if parser == nil {
		parser = BuildParser()
	}

	// Read file
	data, err := ioutil.ReadFile(filename)
	check(err)
	dataString := string(data)

	// Parse the config
	err = parser.ParseString(dataString, config)
	checkFileError(err, filename)

	return
}

func isValidExpansionMacro(str string) bool {
	re := regexp.MustCompile("^%{[\\w_]+(,[\\w_]+)*?}$") // Permissive
	indices := re.FindAllStringIndex(str, -1)

	if indices != nil && indices[0][0] == 0 && indices[0][1] == len(str) {
		// "%" is the first and last value, and they contain stuff
		return true
	}

	return false
}

func splitIdentifiers(identStr string) (idents []string) {
	if len(identStr) > 0 && identStr[:1][0] == '.' { // Remove leading dot
		identStr = identStr[1:]
	}

	return strings.Split(identStr, ".")
}

func splitExpansionMacro(macroStr string) []string {
	return strings.Split(macroStr[2:len(macroStr)-1], ",") // Remove leading "%{" and trailing "}" and then split string at comma
}
