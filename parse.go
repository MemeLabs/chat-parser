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

func NewParserContext(opt ParserContextValues) *sparserContext {
	return &sparserContext{
		emotes:         toMap(opt.Emotes),
		emoteModifiers: toMap(opt.EmoteModifiers),
		nicks:          toMap(opt.Nicks),
		tags:           toMap(opt.Tags),
	}
}

type sparserContext struct {
	emotes         map[string]struct{}
	emoteModifiers map[string]struct{}
	nicks          map[string]struct{}
	tags           map[string]struct{}
}

func NewParser(ctx *sparserContext, tokens []sitem) *sparser {
	return &sparser{
		ctx:    ctx,
		tokens: tokens,
		i:      -1,
	}
}

type sparser struct {
	ctx *sparserContext

	tokens []sitem
	i      int

	pos int
	tok sitemType
	lit string
}

func (p *sparser) next() {
	p.i++
	t := p.tokens[p.i]
	p.tok = t.typ
	p.pos = t.pos
	p.lit = string(t.val)
}

func (p *sparser) parseEmote() (e *Emote) {
	e = &Emote{
		Name:   p.lit,
		TokPos: p.pos,
	}

	for {
		p.next()
		e.TokEnd = p.pos

		if p.lit != ":" {
			return
		}
		p.next()

		if _, ok := p.ctx.emoteModifiers[p.lit]; !ok {
			return
		}
		e.InsertModifier(p.lit)
	}
}

func (p *sparser) parseTag() (e *Tag) {
	pos := p.pos
	name := p.lit

	p.next()

	return &Tag{
		Name:   name,
		TokPos: pos,
		TokEnd: p.pos,
	}
}

func (p *sparser) parseNick() (e *Nick) {
	pos := p.pos
	nick := p.lit

	p.next()

	return &Nick{
		Nick:   nick,
		TokPos: pos,
		TokEnd: p.pos,
	}
}

func (p *sparser) tryParseAtNick() (e *Nick) {
	pos := p.pos

	p.next()

	if _, ok := p.ctx.nicks[p.lit]; !ok {
		return
	}

	e = p.parseNick()
	e.TokPos = pos
	return
}

func (p *sparser) parseCode() (s *Span) {
	pos := p.pos

	for p.tok != sitemEOF {
		p.next()
		if p.lit == "`" {
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

func (p *sparser) parseSpan(t SpanType) (s *Span) {
	s = &Span{
		Type:   t,
		TokPos: p.pos,
	}

	p.next()

	if p.lit == ">" && t == SpanMessage {
		s.Type = SpanGreentext
		p.next()
	}

	for {
		switch p.tok {
		case sitemEOF:
			s.TokEnd = p.pos
			return
		case sitemSpoiler:
			if t == SpanSpoiler {
				p.next()
				s.TokEnd = p.pos
				return
			}
			s.Insert(p.parseSpan(SpanSpoiler))
		case sitemPunct:
			if p.lit == "`" {
				s.Insert(p.parseCode())
			} else if p.lit == "@" {
				if n := p.tryParseAtNick(); n != nil {
					s.Insert(n)
				}
			} else {
				p.next()
			}
		case sitemWord:
			if _, ok := p.ctx.tags[p.lit]; ok {
				s.Insert(p.parseTag())
			} else if _, ok := p.ctx.emotes[p.lit]; ok {
				s.Insert(p.parseEmote())
			} else if _, ok := p.ctx.nicks[p.lit]; ok {
				s.Insert(p.parseNick())
			} else {
				p.next()
			}
		case sitemWhitespace:
			p.next()
		}
	}
}

func (p *sparser) parseMessage() (s *Span) {
	return p.parseSpan(SpanMessage)
}
