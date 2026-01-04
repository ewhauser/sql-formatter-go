package sqlformatter

// InlineLayout formats single-line expressions with width limit.
type InlineLayout struct {
	*Layout
	expressionWidth int
	length          int
	trailingSpace   bool
}

type InlineLayoutError struct{}

func (e InlineLayoutError) Error() string { return "inline layout error" }

func NewInlineLayout(expressionWidth int) *InlineLayout {
	return &InlineLayout{Layout: NewLayout(NewIndentation("")), expressionWidth: expressionWidth}
}

func (l *InlineLayout) Add(items ...interface{}) {
	for _, item := range items {
		l.addToLength(item)
		if l.length > l.expressionWidth {
			panic(InlineLayoutError{})
		}
	}
	l.Layout.Add(items...)
}

func (l *InlineLayout) addToLength(item interface{}) {
	switch v := item.(type) {
	case string:
		l.length += len(v)
		l.trailingSpace = false
	case WS:
		switch v {
		case MandatoryNewline, Newline:
			panic(InlineLayoutError{})
		case Indent, SingleIndent, Space:
			if !l.trailingSpace {
				l.length++
				l.trailingSpace = true
			}
		case NoNewline, NoSpace:
			if l.trailingSpace {
				l.trailingSpace = false
				l.length--
			}
		}
	}
}
