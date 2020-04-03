package parser

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError:              "error",
	itemEmote:              "emote",
	itemEmoteModifier:      "emote modifier",
	itemEmoteModifierDelim: "emote modifier delim",
	itemSpoilerDelim:       "spoiler delim",
	itemSpoilerText:        "spoiler",
	itemCode:               "code",
	itemCodeDelim:          "code delim",
	itemLink:               "link",
	itemText:               "text",
	itemEOF:                "eof",
	itemGreenText:          "greentext",
	itemUsername:           "username",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
	ast   *Span
}

func mkItem(typ itemType, text string) item {
	return item{
		typ: typ,
		val: text,
	}
}

var (
	tCodeDelim     = mkItem(itemCodeDelim, "`")
	tEOF           = mkItem(itemEOF, "")
	tSpoilerDelim  = mkItem(itemSpoilerDelim, "||")
	tEmoteModDelim = mkItem(itemEmoteModifierDelim, ":")
)

var lexTests = []lexTest{
	{"at without username in spoiler", "||`||||@||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "`"),
		tSpoilerDelim,
		tSpoilerDelim,
		mkItem(itemSpoilerText, "@"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"emote with trailing", "PEPE0", []item{
		mkItem(itemText, "PEPE0"),
		tEOF,
	}, &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 5,
	}},
	{"text with code", "text `with code`", []item{
		mkItem(itemText, "text "),
		tCodeDelim,
		mkItem(itemCode, "with code"),
		tCodeDelim,
		tEOF,
	}, &Span{
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
	{"just code", "`just code`", []item{
		tCodeDelim,
		mkItem(itemCode, "just code"),
		tCodeDelim,
		tEOF,
	}, &Span{
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
	{"unclosed code tag", "text `code?", []item{
		mkItem(itemText, "text `code?"),
		tEOF,
	}, &Span{
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
	{"avoid out of range", "text `", []item{
		mkItem(itemText, "text `"),
		tEOF,
	}, &Span{
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
	{"just text", "why even test this case?", []item{
		mkItem(itemText, "why even test this case?"),
		tEOF,
	}, &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 24,
	}},
	{"text and spoiler", "text ||and a spoiler||", []item{
		mkItem(itemText, "text "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "and a spoiler"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"justspoiler", "||spoiler||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"code and spoiler", "`code` and ||spoiler||", []item{
		tCodeDelim,
		mkItem(itemCode, "code"),
		tCodeDelim,
		mkItem(itemText, " and "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"spoiler and code", "||spoiler|| and `code`", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		mkItem(itemText, " and "),
		tCodeDelim,
		mkItem(itemCode, "code"),
		tCodeDelim,
		tEOF,
	}, &Span{
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
	{"empty code", "``", []item{
		tCodeDelim,
		tCodeDelim,
		tEOF,
	}, &Span{
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
	{"empty spoiler", "||||", []item{
		tSpoilerDelim,
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"spoiler out of range", "|", []item{
		mkItem(itemText, "|"),
		tEOF,
	}, &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 1,
	}},
	{"spoiler meme", "|||", []item{
		mkItem(itemText, "|||"),
		tEOF,
	}, &Span{
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
	{"just emote", "PEPE", []item{
		mkItem(itemEmote, "PEPE"),
		tEOF,
	}, &Span{
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
	{"text and emote", "haha PEPE test", []item{
		mkItem(itemText, "haha "),
		mkItem(itemEmote, "PEPE"),
		mkItem(itemText, " test"),
		tEOF,
	}, &Span{
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
	{"emote with modifier", "PEPE:wide", []item{
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		tEOF,
	}, &Span{
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
	{"text and emote", "haha PEPE:wide test", []item{
		mkItem(itemText, "haha "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		mkItem(itemText, " test"),
		tEOF,
	}, &Span{
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
	{"emote in spoiler", "test ||spoiler PEPE ||", []item{
		mkItem(itemText, "test "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		mkItem(itemSpoilerText, " "),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"emote in spoiler", "test ||spoiler PEPE||", []item{
		mkItem(itemText, "test "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"emote in spoiler with mod", "||spoiler PEPE:wide||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"emotewith mod in middle of spoiler", "||spoiler PEPE:wide spoiler||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		mkItem(itemSpoilerText, " spoiler"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"uneven spoiler", "test ||spoiler uneven", []item{
		mkItem(itemText, "test ||spoiler uneven"),
		tEOF,
	}, &Span{
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
	{"uneven code", "test `spoiler uneven", []item{
		mkItem(itemText, "test `spoiler uneven"),
		tEOF,
	}, &Span{
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
	{"lots of stuff", "text and `code PEPE` and maybe ||a spoiler PEPE:wide CuckCrab|| `...`", []item{
		mkItem(itemText, "text and "),
		tCodeDelim,
		mkItem(itemCode, "code PEPE"),
		tCodeDelim,
		mkItem(itemText, " and maybe "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "a spoiler "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		mkItem(itemSpoilerText, " "),
		mkItem(itemEmote, "CuckCrab"),
		tSpoilerDelim,
		mkItem(itemText, " "),
		tCodeDelim,
		mkItem(itemCode, "..."),
		tCodeDelim,
		tEOF,
	}, &Span{
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
	{"whater this is", "`||`||`Abathur:flip `||", []item{
		tCodeDelim,
		mkItem(itemCode, "||"),
		tCodeDelim,
		tSpoilerDelim,
		tCodeDelim,
		mkItem(itemCode, "Abathur:flip "),
		tCodeDelim,
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"greentext", ">implying this lexer works", []item{
		mkItem(itemGreenText, ">implying this lexer works"),
		tEOF,
	}, &Span{
		Type:   SpanGreentext,
		TokPos: 0,
		TokEnd: 26,
	}},
	{"greentext", "text >greentext ||spoiler|| greentext agane", []item{
		mkItem(itemText, "text "),
		mkItem(itemGreenText, ">greentext "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		mkItem(itemGreenText, " greentext agane"),
		tEOF,
	}, &Span{
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
	{"greentext", "text >greentext ||spoiler|| PEPE CuckCrab:spin greentext `code` agane", []item{
		mkItem(itemText, "text "),
		mkItem(itemGreenText, ">greentext "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		mkItem(itemGreenText, " "),
		mkItem(itemEmote, "PEPE"),
		mkItem(itemGreenText, " "),
		mkItem(itemEmote, "CuckCrab"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "spin"),
		mkItem(itemGreenText, " greentext "),
		tCodeDelim,
		mkItem(itemCode, "code"),
		tCodeDelim,
		mkItem(itemGreenText, " agane"),
		tEOF,
	}, &Span{
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
	{"username", "jeanpierrepratt hi", []item{
		mkItem(itemUsername, "jeanpierrepratt"),
		mkItem(itemText, " hi"),
		tEOF,
	}, &Span{
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
	{"username", "@abeous hi", []item{
		mkItem(itemUsername, "@abeous"),
		mkItem(itemText, " hi"),
		tEOF,
	}, &Span{
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
	{"username in spoiler", "hi ||@wrxst||", []item{
		mkItem(itemText, "hi "),
		tSpoilerDelim,
		mkItem(itemUsername, "@wrxst"),
		tSpoilerDelim,
		tEOF,
	}, &Span{
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
	{"url", "https://unicode.org/reports/tr44/#Grapheme_Extend", []item{
		tEOF,
	}, nil},
	{"emoji", "🙈🙉🙊", []item{
		tEOF,
	}, &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 3,
	}},
	{"non ascii words", "日本語のテキスト", []item{
		tEOF,
	}, &Span{
		Type:   SpanMessage,
		TokPos: 0,
		TokEnd: 8,
	}},
	{"code spoiler mashup", "||`||`", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "`"),
		tSpoilerDelim,
		mkItem(itemText, "`"),
		tEOF,
	}, &Span{
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
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens := slex(test.input)

		ctx := NewParserContext(ParserContextValues{
			Emotes:         []string{"PEPE", "CuckCrab"},
			EmoteModifiers: []string{"wide", "rustle", "spin"},
			Nicks:          []string{"abeous", "jeanpierrepratt", "wrxst"},
			Tags:           []string{"nsfw"},
		})
		p := NewParser(ctx, tokens)

		ast := p.parseMessage()

		if test.ast != nil {
			if !reflect.DeepEqual(test.ast, ast) {
				t.Errorf("%s: got\n%s\nexpected\n%s", test.name, spew.Sdump(ast), spew.Sdump(test.ast))
			}
		} else {
			log.Println(spew.Sdump(ast))
		}

		// items := collect(&test)
		// if !equal(items, test.items, false) {
		// 	t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		// }
	}
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
		if checkPos && i1[k].line != i2[k].line {
			return false
		}
	}
	return true
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input, []string{"PEPE", "CuckCrab"}, []string{"wide", "rustle", "spin"}, []string{"abeous", "jeanpierrepratt", "wrxst"})
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}
