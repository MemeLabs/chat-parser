package parser

type ParserContextValues struct {
	Emotes         []string
	EmoteModifiers []string
	Nicks          []string
	Tags           []string
}

func toMap(arr []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range arr {
		m[v] = struct{}{}
	}
	return m
}

func NewParserContext(opt ParserContextValues) *parserContext {
	return &parserContext{
		emotes:         toMap(opt.Emotes),
		emoteModifiers: toMap(opt.EmoteModifiers),
		nicks:          toMap(opt.Nicks),
		tags:           toMap(opt.Tags),
	}
}

type parserContext struct {
	emotes         map[string]struct{}
	emoteModifiers map[string]struct{}
	nicks          map[string]struct{}
	tags           map[string]struct{}
}

func NewParser(ctx *parserContext, tokens []token) *parser {
	return &parser{
		ctx:    ctx,
		tokens: tokens,
		i:      -1,
	}
}

type parser struct {
	ctx *parserContext

	tokens []token
	i      int

	pos int
	tok tokType
	lit string
}

func (p *parser) next() {
	p.i++
	t := p.tokens[p.i]
	p.tok = t.typ
	p.pos = t.pos
	p.lit = string(t.val)
}

func (p *parser) parseEmote() (e *Emote) {
	e = &Emote{
		Name:   p.lit,
		TokPos: p.pos,
	}

	for {
		p.next()
		e.TokEnd = p.pos

		if p.tok != tokColon {
			return
		}
		p.next()

		if _, ok := p.ctx.emoteModifiers[p.lit]; !ok {
			return
		}
		e.InsertModifier(p.lit)
	}
}

func (p *parser) parseTag() (e *Tag) {
	pos := p.pos
	name := p.lit

	p.next()

	return &Tag{
		Name:   name,
		TokPos: pos,
		TokEnd: p.pos,
	}
}

func (p *parser) parseNick() (e *Nick) {
	pos := p.pos
	nick := p.lit

	p.next()

	return &Nick{
		Nick:   nick,
		TokPos: pos,
		TokEnd: p.pos,
	}
}

func (p *parser) tryParseAtNick() (e *Nick) {
	pos := p.pos

	p.next()

	if _, ok := p.ctx.nicks[p.lit]; !ok {
		return
	}

	e = p.parseNick()
	e.TokPos = pos
	return
}

func (p *parser) parseCode() (s *Span) {
	pos := p.pos

	for p.tok != tokEOF {
		p.next()
		if p.tok == tokBacktick {
			p.next()
			break
		}
	}

	return &Span{
		Type:   SpanCode,
		TokPos: pos,
		TokEnd: p.pos,
	}
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
			} else {
				p.next()
			}
		case tokWord:
			if _, ok := p.ctx.tags[p.lit]; ok {
				s.Insert(p.parseTag())
			} else if _, ok := p.ctx.emotes[p.lit]; ok {
				s.Insert(p.parseEmote())
			} else if _, ok := p.ctx.nicks[p.lit]; ok {
				s.Insert(p.parseNick())
			} else {
				p.next()
			}
		default:
			p.next()
		}
	}
}

func (p *parser) parseMessage() (s *Span) {
	return p.parseSpan(SpanMessage)
}
