package parser

import (
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type parseTest struct {
	name  string
	input string
	ast   *Span
}

var parseTests = []parseTest{
	{"at without username in spoiler", "||`||||@||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Span{
						Type:   SpanCode,
						TokPos: 2,
						TokEnd: 10,
					},
				},
				TokPos: 0,
				TokEnd: 10,
			},
		},
		TokPos: 0,
		TokEnd: 10,
	}},
	{"emote with trailing", "PEPE0", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 5,
	}},
	{"text with code", "text `with code`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 5,
				TokEnd: 16,
			},
		},
		TokPos: 0,
		TokEnd: 16,
	}},
	{"just code", "`just code`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 0,
				TokEnd: 11,
			},
		},
		TokPos: 0,
		TokEnd: 11,
	}},
	{"unclosed code tag", "text `code?", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 5,
				TokEnd: 11,
			},
		},
		TokPos: 0,
		TokEnd: 11,
	}},
	{"avoid out of range", "text `", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 5,
				TokEnd: 6,
			},
		},
		TokPos: 0,
		TokEnd: 6,
	}},
	{"just text", "why even test this case?", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 24,
	}},
	{"text and spoiler", "text ||and a spoiler||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 5,
				TokEnd: 22,
			},
		},
		TokPos: 0,
		TokEnd: 22,
	}},
	{"justspoiler", "||spoiler||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 0,
				TokEnd: 11,
			},
		},
		TokPos: 0,
		TokEnd: 11,
	}},
	{"code and spoiler", "`code` and ||spoiler||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 0,
				TokEnd: 6,
			},
			&Span{
				Type:   SpanSpoiler,
				TokPos: 11,
				TokEnd: 22,
			},
		},
		TokPos: 0,
		TokEnd: 22,
	}},
	{"spoiler and code", "||spoiler|| and `code`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 0,
				TokEnd: 11,
			},
			&Span{
				Type:   SpanCode,
				TokPos: 16,
				TokEnd: 22,
			},
		},
		TokPos: 0,
		TokEnd: 22,
	}},
	{"empty code", "``", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 0,
				TokEnd: 2,
			},
		},
		TokPos: 0,
		TokEnd: 2,
	}},
	{"empty spoiler", "||||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 0,
				TokEnd: 4,
			},
		},
		TokPos: 0,
		TokEnd: 4,
	}},
	{"spoiler out of range", "|", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 1,
	}},
	{"spoiler meme", "|||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 0,
				TokEnd: 3,
			},
		},
		TokPos: 0,
		TokEnd: 3,
	}},
	{"just emote", "PEPE", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Emote{
				Name:   "PEPE",
				TokPos: 0,
				TokEnd: 4,
			},
		},
		TokPos: 0,
		TokEnd: 4,
	}},
	{"text and emote", "haha PEPE test", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Emote{
				Name:   "PEPE",
				TokPos: 5,
				TokEnd: 9,
			},
		},
		TokPos: 0,
		TokEnd: 14,
	}},
	{"emote with modifier", "PEPE:wide", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Emote{
				Name: "PEPE",
				Modifiers: []string{
					"wide",
				},
				TokPos: 0,
				TokEnd: 9,
			},
		},
		TokPos: 0,
		TokEnd: 9,
	}},
	{"text and emote", "haha PEPE:wide test", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Emote{
				Name: "PEPE",
				Modifiers: []string{
					"wide",
				},
				TokPos: 5,
				TokEnd: 14,
			},
		},
		TokPos: 0,
		TokEnd: 19,
	}},
	{"emote in spoiler", "test ||spoiler PEPE ||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Emote{
						Name:   "PEPE",
						TokPos: 15,
						TokEnd: 19,
					},
				},
				TokPos: 5,
				TokEnd: 22,
			},
		},
		TokPos: 0,
		TokEnd: 22,
	}},
	{"emote in spoiler", "test ||spoiler PEPE||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Emote{
						Name:   "PEPE",
						TokPos: 15,
						TokEnd: 19,
					},
				},
				TokPos: 5,
				TokEnd: 21,
			},
		},
		TokPos: 0,
		TokEnd: 21,
	}},
	{"emote in spoiler with mod", "||spoiler PEPE:wide||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Emote{
						Name: "PEPE",
						Modifiers: []string{
							"wide",
						},
						TokPos: 10,
						TokEnd: 19,
					},
				},
				TokPos: 0,
				TokEnd: 21,
			},
		},
		TokPos: 0,
		TokEnd: 21,
	}},
	{"emotewith mod in middle of spoiler", "||spoiler PEPE:wide spoiler||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Emote{
						Name: "PEPE",
						Modifiers: []string{
							"wide",
						},
						TokPos: 10,
						TokEnd: 19,
					},
				},
				TokPos: 0,
				TokEnd: 29,
			},
		},
		TokPos: 0,
		TokEnd: 29,
	}},
	{"uneven spoiler", "test ||spoiler uneven", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 5,
				TokEnd: 21,
			},
		},
		TokPos: 0,
		TokEnd: 21,
	}},
	{"uneven code", "test `spoiler uneven", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 5,
				TokEnd: 20,
			},
		},
		TokPos: 0,
		TokEnd: 20,
	}},
	{"lots of stuff", "text and `code PEPE` and maybe ||a spoiler PEPE:wide CuckCrab|| `...`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 9,
				TokEnd: 20,
			},
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Emote{
						Name: "PEPE",
						Modifiers: []string{
							"wide",
						},
						TokPos: 43,
						TokEnd: 52,
					},
					&Emote{
						Name:   "CuckCrab",
						TokPos: 53,
						TokEnd: 61,
					},
				},
				TokPos: 31,
				TokEnd: 63,
			},
			&Span{
				Type:   SpanCode,
				TokPos: 64,
				TokEnd: 69,
			},
		},
		TokPos: 0,
		TokEnd: 69,
	}},
	{"whater this is", "`||`||`Abathur:flip `||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 0,
				TokEnd: 4,
			},
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Span{
						Type:   SpanCode,
						TokPos: 6,
						TokEnd: 21,
					},
				},
				TokPos: 4,
				TokEnd: 23,
			},
		},
		TokPos: 0,
		TokEnd: 23,
	}},
	{"greentext", ">implying this lexer works", &Span{
		Type:   SpanGreentext,
		TokPos: 0,
		TokEnd: 26,
	}},
	{"greentext", "text >greentext ||spoiler|| greentext agane", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 16,
				TokEnd: 27,
			},
		},
		TokPos: 0,
		TokEnd: 43,
	}},
	{"greentext", "text >greentext ||spoiler|| PEPE CuckCrab:spin greentext `code` agane", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanSpoiler,
				TokPos: 16,
				TokEnd: 27,
			},
			&Emote{
				Name:   "PEPE",
				TokPos: 28,
				TokEnd: 32,
			},
			&Emote{
				Name: "CuckCrab",
				Modifiers: []string{
					"spin",
				},
				TokPos: 33,
				TokEnd: 46,
			},
			&Span{
				Type:   SpanCode,
				TokPos: 57,
				TokEnd: 63,
			},
		},
		TokPos: 0,
		TokEnd: 69,
	}},
	{"username", "jeanpierrepratt hi", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Nick{
				Nick:   "jeanpierrepratt",
				TokPos: 0,
				TokEnd: 15,
			},
		},
		TokPos: 0,
		TokEnd: 18,
	}},
	{"username", "@abeous hi", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Nick{
				Nick:   "abeous",
				TokPos: 0,
				TokEnd: 7,
			},
		},
		TokPos: 0,
		TokEnd: 10,
	}},
	{"incorrectly capitalized username", "@ABEOUS hi", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Nick{
				Nick:   "abeous",
				TokPos: 0,
				TokEnd: 7,
			},
		},
		TokPos: 0,
		TokEnd: 10,
	}},
	{"username in spoiler", "hi ||@wrxst||", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Nick{
						Nick:   "wrxst",
						TokPos: 5,
						TokEnd: 11,
					},
				},
				TokPos: 3,
				TokEnd: 13,
			},
		},
		TokPos: 0,
		TokEnd: 13,
	}},
	{"emoji", "🙈🙉🙊", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 3,
	}},
	{"non ascii words", "日本語のテキスト", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 8,
	}},
	{"more non ascii chars", "Ǆ؁‱ஹ௸௵꧄.ဪ꧅⸻𒈙𒐫﷽", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 14,
	}},
	{"code spoiler mashup", "||`||`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type: SpanSpoiler,
				Nodes: []Node{
					&Span{
						Type:   SpanCode,
						TokPos: 2,
						TokEnd: 6,
					},
				},
				TokPos: 0,
				TokEnd: 6,
			},
		},
		TokPos: 0,
		TokEnd: 6,
	}},
	{"at", "@", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 1,
	}},
	{"me", "/me test", &Span{
		Type:   SpanMe,
		TokPos: 4,
		TokEnd: 8,
	}},
	{"me with multiple spaces", "/me    test", &Span{
		Type:   SpanMe,
		TokPos: 7,
		TokEnd: 11,
	}},
	{"escape sequences", "\\` test `co\\`de`", &Span{
		Type: SpanMessage,
		Nodes: []Node{
			&Span{
				Type:   SpanCode,
				TokPos: 8,
				TokEnd: 16,
			},
		},
		TokPos: 0,
		TokEnd: 16,
	}},
	{"backslash", "\\", &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 1,
	}},
}

func TestParse(t *testing.T) {
	ctx := NewParserContext(ParserContextValues{
		Emotes:         []string{"PEPE", "CuckCrab"},
		EmoteModifiers: []string{"wide", "rustle", "spin"},
		Nicks:          []string{"abeous", "jeanpierrepratt", "wrxst"},
		Tags:           []string{"nsfw"},
	})

	for _, test := range parseTests {
		p := NewParser(ctx, NewLexer(test.input))
		ast := p.ParseMessage()

		if test.ast != nil {
			if !reflect.DeepEqual(test.ast, ast) {
				t.Errorf("%s: got\n%s\nexpected\n%s", test.name, spew.Sdump(ast), spew.Sdump(test.ast))
			}
		} else {
			log.Println(spew.Sdump(ast))
		}
	}
}

func TestRuneIndex(t *testing.T) {
	v := NewRuneIndex(RunesFromStrings([]string{"g", "d", "a", "c", "f"}))

	expected := [][]rune{{'a'}, {'c'}, {'d'}, {'f'}, {'g'}}
	if !reflect.DeepEqual(expected, v.values) {
		t.Error("new rune index should be sorted")
		t.FailNow()
	}

	v.Insert([]rune("b"))
	v.Insert([]rune("e"))

	expected = [][]rune{{'a'}, {'b'}, {'c'}, {'d'}, {'e'}, {'f'}, {'g'}}
	if !reflect.DeepEqual(expected, v.values) {
		t.Error("rune index should remain sorted after inserting values")
		t.FailNow()
	}

	v.Remove([]rune("c"))
	v.Remove([]rune("f"))

	expected = [][]rune{{'a'}, {'b'}, {'d'}, {'e'}, {'g'}}
	if !reflect.DeepEqual(expected, v.values) {
		t.Error("rune index should remain sorted after inserting values")
		t.FailNow()
	}
}

func TestNickIndex(t *testing.T) {
	v := NewNickIndex(RunesFromStrings([]string{
		"FOO",
		"Bar",
		"baz",
	}))

	cases := []struct {
		nick     []rune
		expected bool
	}{
		{[]rune{'f', 'o', 'o'}, true},
		{[]rune{'B', 'A', 'R'}, true},
		{[]rune{'q', 'u', 'x'}, false},
	}

	for _, c := range cases {
		if v.Contains(c.nick) != c.expected {
			t.Errorf("v.Contains '%s' expected %t", string(c.nick), c.expected)
			t.FailNow()
		}
	}
}

func BenchmarkParse(b *testing.B) {
	ctx := NewParserContext(ParserContextValues{
		Emotes:         []string{"PEPE", "CuckCrab"},
		EmoteModifiers: []string{"wide", "rustle", "spin"},
		Nicks:          []string{"abeous", "jeanpierrepratt", "wrxst"},
		Tags:           []string{"nsfw"},
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		test := parseTests[i%len(parseTests)]
		p := NewParser(ctx, NewLexer(test.input))
		ast := p.ParseMessage()
		_ = ast
	}
}

func BenchmarkParseCorpus(b *testing.B) {
	samples := make([]string, 300)
	for i := 0; i < len(samples); i++ {
		d, _ := ioutil.ReadFile(path.Join(".", "corpus", strconv.Itoa(i)))
		samples[i] = string(d)
	}

	ctx := NewParserContext(ParserContextValues{
		Emotes:         []string{"PEPE", "CuckCrab"},
		EmoteModifiers: []string{"wide", "rustle", "spin"},
		Nicks:          []string{"abeous", "jeanpierrepratt", "wrxst"},
		Tags:           []string{"nsfw"},
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := NewParser(ctx, NewLexer(samples[i%len(samples)]))
		ast := p.ParseMessage()
		_ = ast
	}
}
