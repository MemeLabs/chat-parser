// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

// Pos represent the position in the input text
type Pos int

func (i item) String() string {
	return fmt.Sprintf("%q (%s)", i.val, i.typ)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError         itemType = iota // error occurred; value is text of error
	itemEmote                         // an emote
	itemEmoteModifier                 // an emote modifier
	itemEmoteModifierDelim
	itemSpoilerDelim // start or end tag of a spoler "||"
	itemSpoilerText
	itemCodeDelim // start or end tag of code "`"
	itemCode
	itemLink // a link
	itemText // plain text
	itemGreenText
	itemUsername
	itemEOF // end of a message
)

var key = map[string]itemType{
	"`":  itemCodeDelim,
	"||": itemSpoilerDelim,
}

const eof = -1

// Trimming spaces.
// If the action begins "{{- " rather than "{{", then all space/tab/newlines
// preceding the action are trimmed; conversely if it ends " -}}" the
// leading spaces are trimmed. This is done entirely in the lexer; the
// parser never sees it happen. We require an ASCII space to be
// present to avoid ambiguity with things like "{{-3}}". It reads
// better with the space present anyway. For simplicity, only ASCII
// space does the job.
const (
	spaceChars = " \t\r\n" // These are the space characters defined by Go itself.
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name           string    // the name of the input; used only for error reports
	input          string    // the string being scanned
	pos            Pos       // current position in the input
	start          Pos       // start position of this item
	width          Pos       // width of last rune read from input
	items          chan item // channel of scanned items
	parenDepth     int       // nesting depth of ( ) exprs
	startLine      int       // start line of this item
	inSpoiler      bool      // true if in spoiler
	inGreen        bool      //true if in greentext
	emotes         []string  // list of emotes
	emoteMidifiers []string  // list of modifiers
	usernames      []string
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(name, input string, emotes, modifiers, usernames []string) *lexer {
	l := &lexer{
		name:           name,
		input:          input,
		items:          make(chan item),
		startLine:      1,
		emotes:         emotes,
		emoteMidifiers: modifiers,
		usernames:      usernames,
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// state functions

const (
	spoiler       = "||"
	code          = "`"
	modifierStart = ":"
)

// lexText scans until an opening action delimiter, "{{".
func lexText(l *lexer) stateFn {
	l.width = 0
	for {
		// scan until we find a code or spoiler delim
		switch l.next() {
		case '>':
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.ignore()
			return lexGreenText
		case '`':
			l.backup()
			if x := strings.Index(l.input[l.pos+1:], code); x >= 0 {
				if l.pos > l.start {
					l.emit(itemText)
				}
				l.ignore()
				return lexOpeningCode
			}
			l.next()
		case '|':
			l.backup()
			if strings.HasPrefix(l.input[l.pos:], spoiler) {
				if x := strings.Index(l.input[l.pos+2:], spoiler); x >= 0 {
					if l.pos > l.start {
						l.emit(itemText)
					}
					l.ignore()
					return lexOpeningSpoiler
				}
			}
			l.next()
		case eof:
			if l.pos > l.start {
				l.emit(itemText)
			}
			l.ignore()
			l.emit(itemEOF)
			return nil
		default:
			// TODO: we only need to check this if the previous rune was whitespace
			if l.scanEmote() {
				if l.pos > l.start {
					l.emit(itemText)
				}
				l.ignore()
				return lexEmote
			}
			if l.scanUsername() {
				if l.pos > l.start {
					l.emit(itemText)
				}
				l.ignore()
				return lexUsername
			}
		}
	}
}

func lexUsername(l *lexer) stateFn {
	l.width = 0
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r) ||
			r == '`' ||
			r == '|' ||
			r == '>' ||
			r == eof:
			l.backup()
			if l.pos > l.start {
				l.emit(itemUsername)
			}
			l.ignore()
			if l.inSpoiler {
				return lexSpoiler
			}
			if l.inGreen {
				return lexGreenText
			}
			return lexText
		}
	}
}

func lexGreenText(l *lexer) stateFn {
	l.inGreen = true
	l.width = 0
	for {
		// scan until we find a code or spoiler delim
		switch l.next() {
		case '`':
			l.backup()
			if x := strings.Index(l.input[l.pos+1:], code); x >= 0 {
				if l.pos > l.start {
					l.emit(itemGreenText)
				}
				l.ignore()
				return lexOpeningCode
			}
			l.next()
		case '|':
			l.backup()
			if strings.HasPrefix(l.input[l.pos:], spoiler) {
				if x := strings.Index(l.input[l.pos+2:], spoiler); x >= 0 {
					if l.pos > l.start {
						l.emit(itemGreenText)
					}
					l.ignore()
					return lexOpeningSpoiler
				}
			}
			l.next()
		case eof:
			if l.pos > l.start {
				l.emit(itemGreenText)
			}
			l.ignore()
			l.emit(itemEOF)
			return nil
		default:
			// TODO: we only need to check this if the previous rune was whitespace
			if l.scanEmote() {
				if l.pos > l.start {
					l.emit(itemGreenText)
				}
				l.ignore()
				return lexEmote
			}
			if l.scanUsername() {
				if l.pos > l.start {
					l.emit(itemGreenText)
				}
				l.ignore()
				return lexUsername
			}
		}
	}
}

func lexEmote(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == ':':
			l.backup()
			l.emit(itemEmote)
			return lexEmoteModifierDelim
		case unicode.IsSpace(r) || r == '|':
			l.backup()
			l.emit(itemEmote)
			// TODO: move to function
			if l.inSpoiler {
				return lexSpoiler
			}
			if l.inGreen {
				return lexGreenText
			}
			return lexText
		case r == eof:
			l.emit(itemEmote)
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexEmoteModifierDelim(l *lexer) stateFn {
	if l.scanEmoteModifier() {
		l.width = 0
		l.pos += Pos(1)
		l.emit(itemEmoteModifierDelim)
		return lexEmoteModifier
	}
	if l.inSpoiler {
		return lexSpoiler
	}
	if l.inGreen {
		return lexGreenText
	}
	return lexText
}

func lexEmoteModifier(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case unicode.IsSpace(r) || r == '|':
			l.backup()
			l.emit(itemEmoteModifier)
			if l.inSpoiler {
				return lexSpoiler
			}
			if l.inGreen {
				return lexGreenText
			}
			return lexText
		case r == eof:
			l.emit(itemEmoteModifier)
			l.emit(itemEOF)
			return nil
		}
	}
}

func lexOpeningSpoiler(l *lexer) stateFn {
	l.width = 0
	l.pos += Pos(2)
	l.emit(itemSpoilerDelim)
	l.inSpoiler = true
	return lexSpoiler
}

func lexClosingSpoiler(l *lexer) stateFn {
	l.width = 0
	l.pos += Pos(2)
	l.emit(itemSpoilerDelim)
	l.inSpoiler = false
	if l.inGreen {
		return lexGreenText
	}
	return lexText
}

func lexSpoiler(l *lexer) stateFn {
	l.width = 0
	for {
		// scan until we find an emote or the end of the spoiler
		switch l.next() {
		case '`':
			l.backup()
			if x := strings.Index(l.input[l.pos+1:], code); x >= 0 {
				if l.pos > l.start {
					l.emit(itemSpoilerText)
				}
				l.ignore()
				return lexOpeningCode
			}
			l.next()
		case '|':
			l.backup()
			if strings.HasPrefix(l.input[l.pos:], spoiler) {
				if l.pos > l.start {
					l.emit(itemSpoilerText)
				}
				l.ignore()
				return lexClosingSpoiler
			}
			l.next()
		default:
			// TODO: we only need to check this if the previous rune was whitespace
			if l.scanEmote() {
				if l.pos > l.start {
					l.emit(itemSpoilerText)
				}
				l.ignore()
				return lexEmote
			}
			if l.scanUsername() {
				if l.pos > l.start {
					l.emit(itemSpoilerText)
				}
				l.ignore()
				return lexUsername
			}
		}
	}
}

func lexOpeningCode(l *lexer) stateFn {
	l.width = 0
	l.pos += Pos(1)
	l.emit(itemCodeDelim)
	return lexCode
}

func lexClosingCode(l *lexer) stateFn {
	l.width = 0
	l.pos += Pos(1)
	l.emit(itemCodeDelim)
	if l.inSpoiler {
		return lexSpoiler
	}
	if l.inGreen {
		return lexGreenText
	}
	return lexText
}

func lexCode(l *lexer) stateFn {
	l.width = 0
	l.pos += Pos(strings.Index(l.input[l.pos:], code))
	if l.pos > l.start {
		l.emit(itemCode)
	}
	l.ignore()
	return lexClosingCode
}

// TODO: can we speed this up?
// returs true if an emote starts at l.pos
func (l *lexer) scanEmoteModifier() bool {
	for _, modifier := range l.emoteMidifiers {
		if strings.HasPrefix(l.input[l.pos+1:], modifier) {
			return true
		}
	}
	return false
}

// TODO: can we speed this up?
// returs true if an emote starts at l.pos
func (l *lexer) scanEmote() bool {
	// TODO: remove this from here
	l.backup()
	for _, emote := range l.emotes {
		if strings.HasPrefix(l.input[l.pos:], emote) {
			if Pos(len(l.input)) <= l.pos+Pos(len(emote)) {
				return true
			}
			afterEmote := rune(l.input[l.pos+Pos(len(emote))])

			return afterEmote == ':' ||
				unicode.IsSpace(afterEmote) ||
				strings.HasPrefix(l.input[l.pos+Pos(len(emote)):], "||")
		}
	}
	l.next()
	return false
}

func (l *lexer) scanUsername() bool {
	l.backup()
	// FeelsPepoMan improve this
	hasAt := l.accept("@")
	for _, name := range l.usernames {
		if strings.HasPrefix(l.input[l.pos:], name) {
			if hasAt {
				l.backup()
			}
			return true
		}
	}
	l.next()
	return false
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

type sitemType int

const (
	sitemEOF sitemType = iota
	sitemSpoiler
	sitemPunct
	sitemWhitespace
	sitemWord
)

var sitemNames = map[sitemType]string{
	sitemEOF:        "EOF",
	sitemSpoiler:    "Spoiler",
	sitemPunct:      "Punct",
	sitemWhitespace: "Whitespace",
	sitemWord:       "Word",
}

func (i sitemType) String() string {
	return sitemNames[i]
}

type sitem struct {
	typ sitemType
	pos int
	val []rune
}

func (i sitem) Pos() int {
	return i.pos
}

func (i sitem) End() int {
	return i.pos + len(i.val)
}

func (i sitem) String() string {
	return fmt.Sprintf("(%s %d %s)", i.typ, i.pos, string(i.val))
}

type slexer struct {
	input      []rune
	items      []sitem
	start, pos int
}

func (l *slexer) next() rune {
	l.pos++
	if l.pos < len(l.input) {
		return l.input[l.pos]
	}
	return eof
}

func (l *slexer) backup() {
	l.pos--
}

func (l *slexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *slexer) accept(test func(r rune) bool) bool {
	if test(l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *slexer) emit(t sitemType) {
	l.items = append(l.items, sitem{
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

func (l *slexer) run() {
	for {
		r := l.next()
		switch r {
		case eof:
			l.backup()
			l.emit(sitemEOF)
			return
		case '|':
			if l.accept(func(r rune) bool { return r == '|' }) {
				l.emit(sitemSpoiler)
				continue
			}
			fallthrough
		default:
			if unicode.IsSpace(r) {
				for l.accept(func(r rune) bool { return r != eof && unicode.IsSpace(r) }) {
				}
				l.emit(sitemWhitespace)
			} else if unicode.IsOneOf(nonWord, r) {
				l.emit(sitemPunct)
			} else {
				for l.accept(func(r rune) bool { return r != eof && !unicode.IsOneOf(nonWord, r) }) {
				}
				l.emit(sitemWord)
			}
		}
	}
}

func slex(input string) []sitem {
	l := slexer{
		pos:   -1,
		input: []rune(input),
	}
	l.run()
	return l.items
}
