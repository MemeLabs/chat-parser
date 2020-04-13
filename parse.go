package parser

import (
	"sort"
	"sync"
)

func NewRuneIndex(values [][]rune) *RuneIndex {
	sort.Sort(runeSlices(values))
	return &RuneIndex{values: values}
}

type RuneIndex struct {
	sync.Mutex
	values [][]rune
}

func (r *RuneIndex) findIndex(v []rune) int {
	var min, mid int
	max := len(r.values)

	for min != max {
		mid = (max + min) >> 1
		if compareRuneSlices(r.values[mid], v) < 0 {
			min = mid + 1
		} else {
			max = mid
		}
	}

	return min
}

func (r *RuneIndex) Contains(v []rune) bool {
	r.Lock()
	defer r.Unlock()

	i := r.findIndex(v)
	return i != len(r.values) && compareRuneSlices(r.values[i], v) == 0
}

func (r *RuneIndex) Insert(v []rune) {
	r.Lock()
	defer r.Unlock()

	i := r.findIndex(v)
	r.values = append(r.values, v)
	if i != len(r.values)-1 {
		copy(r.values[i+1:], r.values[i:])
		r.values[i] = v
	}
}

func (r *RuneIndex) Remove(v []rune) {
	r.Lock()
	defer r.Unlock()

	i := r.findIndex(v)
	if i != len(r.values) {
		copy(r.values[i:], r.values[i+1:])
		r.values = r.values[:len(r.values)-1]
	}
}

func (r *RuneIndex) Replace(values [][]rune) {
	sort.Sort(runeSlices(values))

	r.Lock()
	defer r.Unlock()

	r.values = values
}

type runeSlices [][]rune

func (a runeSlices) Len() int           { return len(a) }
func (a runeSlices) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a runeSlices) Less(i, j int) bool { return compareRuneSlices(a[i], a[j]) < 0 }

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

func RunesFromStrings(s []string) (r [][]rune) {
	r = make([][]rune, len(s))
	for i, v := range s {
		r[i] = []rune(v)
	}
	return
}

type ParserContextValues struct {
	Emotes         []string
	EmoteModifiers []string
	Nicks          []string
	Tags           []string
}

func NewParserContext(opt ParserContextValues) *ParserContext {
	return &ParserContext{
		Emotes:         NewRuneIndex(RunesFromStrings(opt.Emotes)),
		EmoteModifiers: NewRuneIndex(RunesFromStrings(opt.EmoteModifiers)),
		Nicks:          NewRuneIndex(RunesFromStrings(opt.Nicks)),
		Tags:           NewRuneIndex(RunesFromStrings(opt.Tags)),
	}
}

type ParserContext struct {
	Emotes         *RuneIndex
	EmoteModifiers *RuneIndex
	Nicks          *RuneIndex
	Tags           *RuneIndex
}

var meCmd = []rune("me")

func NewParser(ctx *ParserContext, l lexer) *Parser {
	return &Parser{
		ctx:   ctx,
		lexer: l,
	}
}

type Parser struct {
	ctx   *ParserContext
	lexer lexer

	pos int
	tok tokType
	lit []rune
}

func (p *Parser) next() {
	t := p.lexer.Next()
	p.tok = t.typ
	p.pos = t.pos
	p.lit = t.val
}

func (p *Parser) parseEmote() (e *Emote) {
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

		if !p.ctx.EmoteModifiers.Contains(p.lit) {
			return
		}
		e.InsertModifier(string(p.lit))
	}
}

func (p *Parser) parseTag() (t *Tag) {
	t = &Tag{
		Name:   string(p.lit),
		TokPos: p.pos,
	}

	p.next()

	t.TokEnd = p.pos
	return
}

func (p *Parser) parseNick() (n *Nick) {
	n = &Nick{
		Nick:   string(p.lit),
		TokPos: p.pos,
	}

	p.next()

	n.TokEnd = p.pos
	return
}

func (p *Parser) tryParseAtNick() (n *Nick) {
	pos := p.pos

	p.next()

	if !p.ctx.Nicks.Contains(p.lit) {
		return
	}

	n = p.parseNick()
	n.TokPos = pos
	return
}

func (p *Parser) parseCode() (s *Span) {
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

func (p *Parser) parseSpan(t SpanType) (s *Span) {
	s = &Span{
		Type:   t,
		TokPos: p.pos,
	}

	p.next()

	if t == SpanMessage {
		switch p.tok {
		case tokRAngle:
			s.Type = SpanGreentext
			p.next()
		case tokRSlash:
			p.next()
			if compareRuneSlices(p.lit, meCmd) == 0 {
				s.Type = SpanMe
				p.next()
				p.next()
				s.TokPos = p.pos
			}
		}
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
			if p.ctx.Tags.Contains(p.lit) {
				s.Insert(p.parseTag())
			} else if p.ctx.Emotes.Contains(p.lit) {
				s.Insert(p.parseEmote())
			} else if p.ctx.Nicks.Contains(p.lit) {
				s.Insert(p.parseNick())
			} else {
				p.next()
			}
		default:
			p.next()
		}
	}
}

func (p *Parser) ParseMessage() (s *Span) {
	return p.parseSpan(SpanMessage)
}
