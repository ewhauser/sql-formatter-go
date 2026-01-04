package sqlformatter

import "fmt"

type ConfigError struct{
	Message string
}

func (e ConfigError) Error() string { return e.Message }

func validateConfig(cfg FormatOptions) (FormatOptions, error) {
	removed := []string{"multilineLists", "newlineBeforeOpenParen", "newlineBeforeCloseParen", "aliasAs", "commaPosition", "tabulateAlias"}
	_ = removed // not applicable in Go API

	if cfg.ExpressionWidth <= 0 {
		return cfg, ConfigError{Message: fmt.Sprintf("expressionWidth config must be positive number. Received %d instead.", cfg.ExpressionWidth)}
	}

	if cfg.Params != nil {
		if !validateParams(cfg.Params) {
			// warning only in JS; ignore here
		}
	}

	if cfg.ParamTypes != nil {
		if !validateParamTypes(cfg.ParamTypes) {
			return cfg, ConfigError{Message: "Empty regex given in custom paramTypes. That would result in matching infinite amount of parameters."}
		}
	}

	return cfg, nil
}

func validateParams(params ParamItemsOrList) bool {
	switch v := params.(type) {
	case []string:
		for _, p := range v {
			if p == "" {
				return false
			}
		}
		return true
	case map[string]string:
		return true
	case ParamItems:
		return true
	default:
		return true
	}
}

func validateParamTypes(paramTypes *ParamTypes) bool {
	if paramTypes == nil {
		return true
	}
	for _, p := range paramTypes.Custom {
		if p.Regex == "" {
			return false
		}
	}
	return true
}
