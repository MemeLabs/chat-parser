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

		modifier := string(p.lit)
		if _, ok := p.ctx.emoteModifiers[modifier]; !ok {
			return
		}
		e.InsertModifier(modifier)
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

	if _, ok := p.ctx.nicks[string(p.lit)]; !ok {
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
			word := string(p.lit)
			if _, ok := p.ctx.tags[word]; ok {
				s.Insert(p.parseTag())
			} else if _, ok := p.ctx.emotes[word]; ok {
				s.Insert(p.parseEmote())
			} else if _, ok := p.ctx.nicks[word]; ok {
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
