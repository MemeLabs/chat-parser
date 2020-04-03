package parser

import (
	"fmt"
	"unicode"
)

const eof rune = -1

type tokType int

const (
	tokEOF tokType = iota
	tokSpoiler
	tokPunct
	tokWhitespace
	tokWord
)

var tokNames = map[tokType]string{
	tokEOF:        "EOF",
	tokSpoiler:    "Spoiler",
	tokPunct:      "Punct",
	tokWhitespace: "Whitespace",
	tokWord:       "Word",
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

type lexer struct {
	input      []rune
	tokens     []token
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

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) accept(test func(r rune) bool) bool {
	if test(l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) emit(t tokType) {
	l.tokens = append(l.tokens, token{
		typ: t,
		pos: l.start,
		val: l.input[l.start : l.pos+1],
	})
	l.start = l.pos + 1
}

var nonWord = []*unicode.RangeTable{
	unicode.Dash,
	unicode.Hyphen,
	unicode.Other_Math,
	unicode.Pattern_Syntax,
	unicode.Pattern_White_Space,
	unicode.Quotation_Mark,
	unicode.Sentence_Terminal,
	unicode.Terminal_Punctuation,
	unicode.White_Space,
}

func (l *lexer) run() {
	for {
		r := l.next()
		switch r {
		case eof:
			l.backup()
			l.emit(tokEOF)
			return
		case '|':
			if l.accept(func(r rune) bool { return r == '|' }) {
				l.emit(tokSpoiler)
				continue
			}
			fallthrough
		default:
			if unicode.IsSpace(r) {
				for l.accept(func(r rune) bool { return r != eof && unicode.IsSpace(r) }) {
				}
				l.emit(tokWhitespace)
			} else if unicode.IsOneOf(nonWord, r) {
				l.emit(tokPunct)
			} else {
				for l.accept(func(r rune) bool { return r != eof && !unicode.IsOneOf(nonWord, r) }) {
				}
				l.emit(tokWord)
			}
		}
	}
}

func lex(input string) []token {
	l := lexer{
		pos:   -1,
		input: []rune(input),
	}
	l.run()
	return l.tokens
}
