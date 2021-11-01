package domain

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/antonmedv/expr"
)

var ErrExperssionParameterNotFound = errors.New("parameter not found")

type Expression string

func (e Expression) String() string {
	return string(e)
}

func (e Expression) EvaluateWithVars(params map[string]interface{}) (interface{}, error) {
	program, err := expr.Compile(e.String())
	if err != nil {
		return nil, fmt.Errorf("invalid expression: %w", err)
	}

	env := make(map[string]interface{})

	for _, c := range program.Constants {
		if reflect.TypeOf(c).Kind() == reflect.String {
			paramKey := reflect.ValueOf(c).String()
			if strings.HasPrefix(paramKey, "$") {
				key := strings.TrimPrefix(paramKey, "$")
				if _, ok := params[key]; !ok {
					return nil, fmt.Errorf("%w: %s", ErrExperssionParameterNotFound, key)
				} else {
					env[paramKey] = params[key]
				}
			}
		}
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf(`evaluating expression "%s": %w`, e, err)
	}
	return result, nil
}
