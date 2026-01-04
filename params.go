package sqlformatter

type ParamItems map[string]string

type ParamItemsOrList interface{}

// Params handles placeholder replacement with given params.
type Params struct {
	params ParamItemsOrList
	index  int
}

func NewParams(params ParamItemsOrList) *Params {
	return &Params{params: params, index: 0}
}

func (p *Params) Get(key string, text string) string {
	if p == nil || p.params == nil {
		return text
	}
	if key != "" {
		if m, ok := p.params.(map[string]string); ok {
			if val, ok := m[key]; ok {
				return val
			}
		}
		// numeric keys might come in as map[string]string too
		if m, ok := p.params.(ParamItems); ok {
			if val, ok := m[key]; ok {
				return val
			}
		}
		if m, ok := p.params.(map[string]interface{}); ok {
			if val, ok := m[key]; ok {
				if s, ok := val.(string); ok {
					return s
				}
			}
		}
		return text
	}

	switch v := p.params.(type) {
	case []string:
		if p.index >= 0 && p.index < len(v) {
			val := v[p.index]
			p.index++
			return val
		}
		p.index++
		return text
	case ParamItems:
		// positional fallback for map with numeric keys
		val, ok := v[intToString(p.index+1)]
		p.index++
		if ok {
			return val
		}
		return text
	case map[string]string:
		val, ok := v[intToString(p.index+1)]
		p.index++
		if ok {
			return val
		}
		return text
	default:
		p.index++
		return text
	}
}

func (p *Params) GetPositionalParameterIndex() int {
	if p == nil {
		return 0
	}
	return p.index
}

func (p *Params) SetPositionalParameterIndex(i int) {
	if p == nil {
		return
	}
	p.index = i
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
