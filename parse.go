package parser

import (
	"sort"
)

type ParserContextValues struct {
	Emotes         []string
	EmoteModifiers []string
	Nicks          []string
	Tags           []string
}

func toRuneSlices(arr []string) (s [][]rune) {
	s = make([][]rune, len(arr))
	for i, v := range arr {
		s[i] = []rune(v)
	}
	sort.Sort(runeSlices(s))
	return
}

func compareRuneSlices(a, b []rune) int {
	if len(a) != len(b) {
		return len(a) - len(b)
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return int(a[i] - b[i])
		}
	}
	return 0
}

func inRuneSlices(r [][]rune, v []rune) bool {
	var min, mid int
	max := len(r)

	for min != max {
		mid = (max + min) >> 1
		if compareRuneSlices(r[mid], v) < 0 {
			min = mid + 1
		} else {
			max = mid
		}
	}

	return min != len(r) && compareRuneSlices(r[min], v) == 0
}

type runeSlices [][]rune

func (a runeSlices) Len() int           { return len(a) }
func (a runeSlices) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a runeSlices) Less(i, j int) bool { return compareRuneSlices(a[i], a[j]) < 0 }

func NewParserContext(opt ParserContextValues) *parserContext {
	return &parserContext{
		emotes:         toRuneSlices(opt.Emotes),
		emoteModifiers: toRuneSlices(opt.EmoteModifiers),
		nicks:          toRuneSlices(opt.Nicks),
		tags:           toRuneSlices(opt.Tags),
	}
}

type parserContext struct {
	emotes         [][]rune
	emoteModifiers [][]rune
	nicks          [][]rune
	tags           [][]rune
}

func NewParser(ctx *parserContext, l lexer) *parser {
	return &parser{
		ctx:   ctx,
		lexer: l,
	}
}

type parser struct {
	ctx   *parserContext
	lexer lexer

	pos int
	tok tokType
	lit []rune
}

func (p *parser) next() {
	t := p.lexer.Next()
	p.tok = t.typ
	p.pos = t.pos
	p.lit = t.val
}

func (p *parser) parseEmote() (e *Emote) {
	e = &Emote{
		Name:   string(p.lit),
		TokPos: p.pos,
	}

	for {
		p.next()
		e.TokEnd = p.pos

		if p.tok != tokColon {
			return
		}
		p.next()

		if !inRuneSlices(p.ctx.emoteModifiers, p.lit) {
			return
		}
		e.InsertModifier(string(p.lit))
	}
}

func (p *parser) parseTag() (t *Tag) {
	t = &Tag{
		Name:   string(p.lit),
		TokPos: p.pos,
	}

	p.next()

	t.TokEnd = p.pos
	return
}

func (p *parser) parseNick() (n *Nick) {
	n = &Nick{
		Nick:   string(p.lit),
		TokPos: p.pos,
	}

	p.next()

	n.TokEnd = p.pos
	return
}

func (p *parser) tryParseAtNick() (n *Nick) {
	pos := p.pos

	p.next()

	if !inRuneSlices(p.ctx.nicks, p.lit) {
		return
	}

	n = p.parseNick()
	n.TokPos = pos
	return
}

func (p *parser) parseCode() (s *Span) {
	s = &Span{
		Type:   SpanCode,
		TokPos: p.pos,
	}

	for p.tok != tokEOF {
		p.next()
		if p.tok == tokBacktick {
			p.next()
			break
		}
	}

	s.TokEnd = p.pos
	return
}

func (p *parser) parseSpan(t SpanType) (s *Span) {
	s = &Span{
		Type:   t,
		TokPos: p.pos,
	}

	p.next()

	if p.tok == tokRAngle && t == SpanMessage {
		s.Type = SpanGreentext
		p.next()
	}

	for {
		switch p.tok {
		case tokEOF:
			s.TokEnd = p.pos
			return
		case tokSpoiler:
			if t == SpanSpoiler {
				p.next()
				s.TokEnd = p.pos
				return
			}
			s.Insert(p.parseSpan(SpanSpoiler))
		case tokBacktick:
			s.Insert(p.parseCode())
		case tokAt:
			if n := p.tryParseAtNick(); n != nil {
				s.Insert(n)
			}
		case tokWord:
			if inRuneSlices(p.ctx.tags, p.lit) {
				s.Insert(p.parseTag())
			} else if inRuneSlices(p.ctx.emotes, p.lit) {
				s.Insert(p.parseEmote())
			} else if inRuneSlices(p.ctx.nicks, p.lit) {
				s.Insert(p.parseNick())
			} else {
				p.next()
			}
		default:
			p.next()
		}
	}
}

func (p *parser) ParseMessage() (s *Span) {
	return p.parseSpan(SpanMessage)
}
