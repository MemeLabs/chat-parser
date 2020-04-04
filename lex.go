package parser

import (
	"fmt"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

const eof rune = -1

type tokType int

const (
	tokEOF tokType = iota
	tokSpoiler
	tokPunct
	tokWhitespace
	tokWord
	tokBacktick
	tokColon
	tokRAngle
	tokAt
)

var tokNames = map[tokType]string{
	tokEOF:        "EOF",
	tokSpoiler:    "Spoiler",
	tokPunct:      "Punct",
	tokWhitespace: "Whitespace",
	tokWord:       "Word",
	tokBacktick:   "Backtick",
	tokColon:      "Colon",
	tokRAngle:     "RAngle",
	tokAt:         "At",
}

func (i tokType) String() string {
	return tokNames[i]
}

type token struct {
	typ tokType
	pos int
	val []rune
}

func (i token) String() string {
	return fmt.Sprintf("(%s %d %s)", i.typ, i.pos, string(i.val))
}

func NewLexer(input string) lexer {
	return lexer{
		input: []rune(input),
		pos:   -1,
	}
}

type lexer struct {
	input      []rune
	start, pos int
}

func (l *lexer) next() rune {
	l.pos++
	if l.pos < len(l.input) {
		return l.input[l.pos]
	}
	return eof
}

func (l *lexer) backup() {
	l.pos--
}

func (l *lexer) accept(test func(r rune) bool) bool {
	if test(l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) emit(t tokType) (tok token) {
	tok = token{
		typ: t,
		pos: l.start,
		val: l.input[l.start : l.pos+1],
	}
	l.start = l.pos + 1
	return
}

var nonWord = rangetable.Merge(
	unicode.Dash,
	unicode.Hyphen,
	unicode.Other_Math,
	unicode.Pattern_Syntax,
	unicode.Pattern_White_Space,
	unicode.Quotation_Mark,
	unicode.Sentence_Terminal,
	unicode.Terminal_Punctuation,
	unicode.White_Space,
)

func (l *lexer) Next() token {
	r := l.next()
	switch r {
	case eof:
		l.backup()
		return l.emit(tokEOF)
	case '`':
		return l.emit(tokBacktick)
	case ':':
		return l.emit(tokColon)
	case '>':
		return l.emit(tokRAngle)
	case '@':
		return l.emit(tokAt)
	case '|':
		if l.accept(func(r rune) bool { return r == '|' }) {
			return l.emit(tokSpoiler)
		} else {
			return l.emit(tokPunct)
		}
	default:
		if unicode.IsSpace(r) {
			for l.accept(func(r rune) bool { return unicode.IsSpace(r) }) {
			}
			return l.emit(tokWhitespace)
		} else if unicode.Is(nonWord, r) {
			return l.emit(tokPunct)
		} else {
			for l.accept(func(r rune) bool { return r != eof && !unicode.Is(nonWord, r) }) {
			}
			return l.emit(tokWord)
		}
	}
}
