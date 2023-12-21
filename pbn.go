package pbn

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

var (
	pbnLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"String", `"(\\"|[^"])*"`},
		{"Contract", `([1-7][CDHSN])|(Pass)|(X)|(XX)`},
		{"Number", `[-+]?(\d*\.)?\d+`},
		{"Name", `([A-Z][a-zA-Z_]*)`},
		{"Ident", `[a-zA-Z_]\w*`},
		{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
		{"NewGame", `\r?\n\s*\r?\n`},
		{"EOL", `[\r?\n]`},
		{"whitespace", `[ \t]+`},
	})

	parser = participle.MustBuild[PBN](
		participle.Lexer(pbnLexer),
	)

	tagParser = participle.MustBuild[Tag](
		participle.Lexer(pbnLexer),
	)

	escapedParser = participle.MustBuild[escapedLine](
		participle.Lexer(pbnLexer),
	)

	tableParser = participle.MustBuild[Table](
		participle.Lexer(pbnLexer),
		participle.Unquote("String"),
	)
)

type PBN struct {
	Game []Game `parser:"@@*"`
}

type Game struct {
	Tags   map[string]Tag
	Tables map[string]Table
}

func (g *Game) Parse(lex *lexer.PeekingLexer) error {
	g.Tags = make(map[string]Tag)
	g.Tables = make(map[string]Table)
	for {
		if lex.Peek().Type == pbnLexer.Symbols()["NewGame"] {
			lex.Next()
			return nil
		}

		_, err := escapedParser.ParseFromLexer(lex, participle.AllowTrailing(true))
		if err == nil {
			continue
		}
		tag, err := tagParser.ParseFromLexer(lex, participle.AllowTrailing(true))
		if err != nil {
			break
		}

		if strings.HasSuffix(tag.Name, "Table") {
			table, err := tableParser.ParseFromLexer(lex, participle.AllowTrailing(true))
			if err != nil {
				if lex.Peek().Type != pbnLexer.Symbols()["NewGame"] {
					continue
				}
			}
			table.Name = tag.Name
			table.ColumnDescription = string(tag.Value)
			g.Tables[tag.Name] = *table

			continue
		}
		if _, exists := g.Tags[tag.Name]; !exists {
			g.Tags[tag.Name] = *tag
		}
	}
	if len(g.Tags) == 0 {
		return participle.NextMatch
	}

	return nil
}

type Tag struct {
	Name  string   `parser:"'[' @Name"`
	Value TagValue `parser:"@String ']' EOL?"`
}

func (t Tag) Equal(other Tag) bool {
	return t.Name == other.Name && t.Value == other.Value
}

type TagValue string

func (v *TagValue) Capture(values []string) error {
	*v = TagValue(strings.Trim(values[0], "\""))
	return nil
}

type escapedLine struct {
	Content string `parser:"'%' @~(EOL | EOF)* (EOL | EOF)"`
}

type Table struct {
	Name              string
	ColumnDescription string
	Data              TableData `parser:"(@((Number | String | '-' | Contract | Ident | Name)+) (EOL | EOF))*"`
}

type TableData [][]string

func (d *TableData) Capture(values []string) error {
	for i := range values {
		values[i] = strings.Trim(values[i], "\"")
	}
	*d = append(*d, values)
	return nil
}
