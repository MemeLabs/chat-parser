package parser

type Node interface {
	Pos() int
	End() int
}

type SpanType int

const (
	SpanMessage SpanType = iota
	SpanText
	SpanCode
	SpanGreentext
	SpanSpoiler
)

var spanTypeNames = map[SpanType]string{
	SpanMessage:   "Message",
	SpanText:      "Text",
	SpanCode:      "Code",
	SpanGreentext: "Greentext",
	SpanSpoiler:   "Spoiler",
}

func (t SpanType) String() string {
	return spanTypeNames[t]
}

type Span struct {
	Type   SpanType
	Nodes  []Node
	TokPos int
	TokEnd int
}

func (s *Span) Insert(n Node) {
	if ns, ok := n.(*Span); ok && ns.Type == s.Type {
		s.Nodes = append(s.Nodes, ns.Nodes...)
		s.TokEnd = ns.TokEnd
	} else {
		s.Nodes = append(s.Nodes, n)
	}
}

func (s *Span) Pos() int {
	return s.TokPos
}

func (s *Span) End() int {
	return s.TokEnd
}

type Link struct {
	URL    string
	TokPos int
	TokEnd int
}

func (l *Link) Pos() int {
	return l.TokPos
}

func (l *Link) End() int {
	return l.TokEnd
}

type Emote struct {
	Name      string
	Modifiers []string
	TokPos    int
	TokEnd    int
}

func (e *Emote) InsertModifier(m string) {
	e.Modifiers = append(e.Modifiers, m)
}

func (e *Emote) Pos() int {
	return e.TokPos
}

func (e *Emote) End() int {
	return e.TokEnd
}

type Tag struct {
	Name   string
	TokPos int
	TokEnd int
}

func (t *Tag) Pos() int {
	return t.TokPos
}

func (t *Tag) End() int {
	return t.TokEnd
}

type Nick struct {
	Nick   string
	TokPos int
	TokEnd int
}

func (n *Nick) Pos() int {
	return n.TokPos
}

func (n *Nick) End() int {
	return n.TokEnd
}
