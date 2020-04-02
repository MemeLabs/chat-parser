// Copyright 2011 The Go Authors. All rights reserved.

// Use of this source code is governed by a BSD-style

// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"testing"
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
	{"text with code", "text `with code`", []item{
		mkItem(itemText, "text "),
		tCodeDelim,
		mkItem(itemCode, "with code"),
		tCodeDelim,
		tEOF,
	}},
	{"just code", "`just code`", []item{
		tCodeDelim,
		mkItem(itemCode, "just code"),
		tCodeDelim,
		tEOF,
	}},
	{"unclosed code tag", "text `code?", []item{
		mkItem(itemText, "text `code?"),
		tEOF,
	}},
	{"avoid out of range", "text `", []item{
		mkItem(itemText, "text `"),
		tEOF,
	}},
	{"just text", "why even test this case?", []item{
		mkItem(itemText, "why even test this case?"),
		tEOF,
	}},
	{"text and spoiler", "text ||and a spoiler||", []item{
		mkItem(itemText, "text "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "and a spoiler"),
		tSpoilerDelim,
		tEOF,
	}},
	{"justspoiler", "||spoiler||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		tEOF,
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
	}},
	{"empty code", "``", []item{
		tCodeDelim,
		tCodeDelim,
		tEOF,
	}},
	{"empty spoiler", "||||", []item{
		tSpoilerDelim,
		tSpoilerDelim,
		tEOF,
	}},
	{"spoiler out of range", "|", []item{
		mkItem(itemText, "|"),
		tEOF,
	}},
	{"spoiler meme", "|||", []item{
		mkItem(itemText, "|||"),
		tEOF,
	}},
	{"just emote", "PEPE", []item{
		mkItem(itemEmote, "PEPE"),
		tEOF,
	}},
	{"text and emote", "haha PEPE test", []item{
		mkItem(itemText, "haha "),
		mkItem(itemEmote, "PEPE"),
		mkItem(itemText, " test"),
		tEOF,
	}},
	{"emote with modifier", "PEPE:wide", []item{
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		tEOF,
	}},
	{"text and emote", "haha PEPE:wide test", []item{
		mkItem(itemText, "haha "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		mkItem(itemText, " test"),
		tEOF,
	}},
	{"emote in spoiler", "test ||spoiler PEPE ||", []item{
		mkItem(itemText, "test "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		mkItem(itemSpoilerText, " "),
		tSpoilerDelim,
		tEOF,
	}},
	{"emote in spoiler", "test ||spoiler PEPE||", []item{
		mkItem(itemText, "test "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		tSpoilerDelim,
		tEOF,
	}},
	{"emote in spoiler with mod", "||spoiler PEPE:wide||", []item{
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler "),
		mkItem(itemEmote, "PEPE"),
		tEmoteModDelim,
		mkItem(itemEmoteModifier, "wide"),
		tSpoilerDelim,
		tEOF,
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
	}},
	{"uneven spoiler", "test ||spoiler uneven", []item{
		mkItem(itemText, "test ||spoiler uneven"),
		tEOF,
	}},
	{"uneven code", "test `spoiler uneven", []item{
		mkItem(itemText, "test `spoiler uneven"),
		tEOF,
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
	}},
	{"greentext", ">implying this lexer works", []item{
		mkItem(itemGreenText, ">implying this lexer works"),
		tEOF,
	}},
	{"greentext", "text >greentext ||spoiler|| greentext agane", []item{
		mkItem(itemText, "text "),
		mkItem(itemGreenText, ">greentext "),
		tSpoilerDelim,
		mkItem(itemSpoilerText, "spoiler"),
		tSpoilerDelim,
		mkItem(itemGreenText, " greentext agane"),
		tEOF,
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
	}},
	{"username", "jeanpierrepratt hi", []item{
		mkItem(itemUsername, "jeanpierrepratt"),
		mkItem(itemText, " hi"),
		tEOF,
	}},
	{"username", "@abeous hi", []item{
		mkItem(itemUsername, "@abeous"),
		mkItem(itemText, " hi"),
		tEOF,
	}},
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
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
	l := lex(t.name, t.input, []string{"PEPE", "CuckCrab"}, []string{"wide", "rustle", "spin"}, []string{"abeous", "jeanpierrepratt"})
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}
