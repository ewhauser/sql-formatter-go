package sqlformatter

import "strings"

// Whitespace modifiers to be used with Layout.Add

type WS int

const (
	Space WS = iota
	NoSpace
	NoNewline
	Newline
	MandatoryNewline
	Indent
	SingleIndent
)

type LayoutItem interface{}

// Layout builds SQL string with whitespace handling.
type Layout struct {
	items       []LayoutItem
	Indentation *Indentation
}

type LayoutWriter interface {
	Add(items ...interface{})
	GetLayoutItems() []LayoutItem
	GetIndentation() *Indentation
	ToString() string
}

func NewLayout(indentation *Indentation) *Layout {
	return &Layout{Indentation: indentation}
}

func (l *Layout) Add(items ...interface{}) {
	for _, item := range items {
		switch v := item.(type) {
		case WS:
			switch v {
			case Space:
				l.items = append(l.items, Space)
			case NoSpace:
				l.trimHorizontalWhitespace()
			case NoNewline:
				l.trimWhitespace()
			case Newline:
				l.trimHorizontalWhitespace()
				l.addNewline(Newline)
			case MandatoryNewline:
				l.trimHorizontalWhitespace()
				l.addNewline(MandatoryNewline)
			case Indent:
				l.addIndentation()
			case SingleIndent:
				l.items = append(l.items, SingleIndent)
			}
		case string:
			l.items = append(l.items, v)
		}
	}
}

func (l *Layout) trimHorizontalWhitespace() {
	for len(l.items) > 0 {
		last := l.items[len(l.items)-1]
		if last == Space || last == SingleIndent {
			l.items = l.items[:len(l.items)-1]
			continue
		}
		break
	}
}

func (l *Layout) trimWhitespace() {
	for len(l.items) > 0 {
		last := l.items[len(l.items)-1]
		if last == Space || last == SingleIndent || last == Newline {
			l.items = l.items[:len(l.items)-1]
			continue
		}
		break
	}
}

func (l *Layout) addNewline(newline WS) {
	if len(l.items) == 0 {
		return
	}
	last := l.items[len(l.items)-1]
	switch last {
	case Newline:
		l.items[len(l.items)-1] = newline
	case MandatoryNewline:
		// keep
	default:
		l.items = append(l.items, newline)
	}
}

func (l *Layout) addIndentation() {
	for i := 0; i < l.Indentation.GetLevel(); i++ {
		l.items = append(l.items, SingleIndent)
	}
}

func (l *Layout) ToString() string {
	var b strings.Builder
	for _, item := range l.items {
		switch v := item.(type) {
		case WS:
			switch v {
			case Space:
				b.WriteString(" ")
			case Newline, MandatoryNewline:
				b.WriteString("\n")
			case SingleIndent:
				b.WriteString(l.Indentation.GetSingleIndent())
			}
		case string:
			b.WriteString(v)
		}
	}
	return b.String()
}

func (l *Layout) GetLayoutItems() []LayoutItem {
	return l.items
}

func (l *Layout) GetIndentation() *Indentation {
	return l.Indentation
}
