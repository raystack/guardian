package evaluator

import (
	"encoding/json"
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

type ExprParam map[string]interface{}

func (m ExprParam) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (e Expression) EvaluateWithStruct(item interface{}) (interface{}, error) {
	result := map[string]interface{}{}

	if item != nil {
		jsonString, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jsonString, &result); err != nil {
			return nil, err
		}
	}

	return e.EvaluateWithVars(result)
}

func (e Expression) EvaluateWithVars(params map[string]interface{}) (interface{}, error) {
	program, err := expr.Compile(e.String())
	if err != nil {
		return nil, fmt.Errorf("invalid expression: %w", err)
	}

	env := make(ExprParam)

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
