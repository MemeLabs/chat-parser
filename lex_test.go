package parser

import (
	"log"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type lexTest struct {
	name  string
	input string
	toks  []token
}

func mkItem(typ tokType, pos int, text string) token {
	return token{
		typ: typ,
		pos: pos,
		val: []rune(text),
	}
}

var lexTests = []lexTest{
	{"at without username in spoiler", "||`||||@||", []token{
		mkItem(tokSpoiler, 0, "||"),
		mkItem(tokBacktick, 2, "`"),
		mkItem(tokSpoiler, 3, "||"),
		mkItem(tokSpoiler, 5, "||"),
		mkItem(tokAt, 7, "@"),
		mkItem(tokSpoiler, 8, "||"),
		mkItem(tokEOF, 10, ""),
	}},
	{"emote with trailing", "PEPE0", []token{
		mkItem(tokWord, 0, "PEPE0"),
		mkItem(tokEOF, 5, ""),
	}},
	{"text with code", "text `with code`", []token{
		mkItem(tokWord, 0, "text"),
		mkItem(tokWhitespace, 4, " "),
		mkItem(tokBacktick, 5, "`"),
		mkItem(tokWord, 6, "with"),
		mkItem(tokWhitespace, 10, " "),
		mkItem(tokWord, 11, "code"),
		mkItem(tokBacktick, 15, "`"),
		mkItem(tokEOF, 16, ""),
	}},
	{"underscores", "words_with_underscores", []token{
		mkItem(tokWord, 0, "words_with_underscores"),
		mkItem(tokEOF, 22, ""),
	}},
	{"emoji", "🙈🙉🙊", []token{
		mkItem(tokWord, 0, "🙈🙉🙊"),
		mkItem(tokEOF, 3, ""),
	}},
	{"non ascii words", "日本語のテキスト", []token{
		mkItem(tokWord, 0, "日本語のテキスト"),
		mkItem(tokEOF, 8, ""),
	}},
	{"more unicode", "Ǆ؁‱ஹ௸௵꧄.ဪ꧅⸻𒈙𒐫﷽", []token{
		mkItem(tokWord, 0, "Ǆ؁"),
		mkItem(tokPunct, 2, "‱"),
		mkItem(tokWord, 3, "ஹ௸௵꧄"),
		mkItem(tokPunct, 7, "."),
		mkItem(tokWord, 8, "ဪ꧅"),
		mkItem(tokPunct, 10, "⸻"),
		mkItem(tokWord, 11, "𒈙𒐫﷽"),
		mkItem(tokEOF, 14, ""),
	}},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		toks := lex(test.input)

		// if test.toks != nil {
		if len(test.toks) != 0 {
			if !reflect.DeepEqual(test.toks, toks) {
				t.Errorf("%s: got\n%s\nexpected\n%s", test.name, spew.Sdump(toks), spew.Sdump(test.toks))
			}
		} else {
			log.Println(spew.Sdump(toks))
		}
	}
}
