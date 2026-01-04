package sqlformatter

const (
	indentTypeTopLevel  = "top-level"
	indentTypeBlockLevel = "block-level"
)

type Indentation struct {
	indent      string
	indentTypes []string
}

func NewIndentation(indent string) *Indentation {
	return &Indentation{indent: indent}
}

func (i *Indentation) GetSingleIndent() string {
	return i.indent
}

func (i *Indentation) GetLevel() int {
	return len(i.indentTypes)
}

func (i *Indentation) IncreaseTopLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeTopLevel)
}

func (i *Indentation) IncreaseBlockLevel() {
	i.indentTypes = append(i.indentTypes, indentTypeBlockLevel)
}

func (i *Indentation) DecreaseTopLevel() {
	if len(i.indentTypes) == 0 {
		return
	}
	if i.indentTypes[len(i.indentTypes)-1] == indentTypeTopLevel {
		i.indentTypes = i.indentTypes[:len(i.indentTypes)-1]
	}
}

func (i *Indentation) DecreaseBlockLevel() {
	for len(i.indentTypes) > 0 {
		last := i.indentTypes[len(i.indentTypes)-1]
		i.indentTypes = i.indentTypes[:len(i.indentTypes)-1]
		if last != indentTypeTopLevel {
			break
		}
	}
}
