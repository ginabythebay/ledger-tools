package rules

import (
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Input struct {
	Key   string
	Value string
}

type Result map[string]string

type RuleSet struct {
	allMappings []mapping
}

func From(config []byte, validInputs, validOutputs []string) (*RuleSet, error) {
	inputSet := map[string]bool{}
	for _, s := range validInputs {
		inputSet[s] = true
	}
	outputSet := map[string]bool{}
	for _, s := range validOutputs {
		outputSet[s] = true
	}

	var parsed []map[string]string
	if err := yaml.Unmarshal(config, &parsed); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	allMappings := []mapping{}

	for _, mp := range parsed {
		var outputKey, outputValue string
		var mr *matcher
		for k, v := range mp {
			if inputSet[k] {
				if mr != nil {
					return nil, errors.Errorf("Unexpected input key %q when we already have %q", k, mr.key)
				}
				mr = &matcher{k, v}
				continue
			}
			if outputSet[k] {
				if outputKey != "" {
					return nil, errors.Errorf("Unexpected output key %q when we already have %q", k, outputKey)
				}
				outputKey = k
				outputValue = v
				continue
			}
			return nil, errors.Errorf("Unexpected key %q, not a valid input or a valid output", k)
		}

		if mr == nil {
			return nil, errors.Errorf("No input key found in %v", mp)
		}
		if outputKey == "" {
			return nil, errors.Errorf("No output key found in %v", mp)
		}

		allMappings = append(allMappings, mapping{*mr, outputKey, outputValue})
	}

	return &RuleSet{allMappings}, nil
}

func (rs *RuleSet) Apply(inputs ...Input) Result {
	result := map[string]string{}

	for _, m := range rs.allMappings {
		if m.matches(inputs) {
			if _, have := result[m.outputKey]; !have {
				result[m.outputKey] = m.outputValue
			}
		}
	}
	return result
}

func (r Result) Get(key string) string {
	return r[key]
}

type mapping struct {
	matcher
	outputKey   string
	outputValue string
}

// Very simple for now.  Exact match only, of one key and one value.
// Maybe ignore case, regex etc, multiple entries, in the future.
type matcher struct {
	key   string
	value string
}

func (m matcher) matches(inputs []Input) bool {
	for _, i := range inputs {
		if i.Key == m.key && i.Value == m.value {
			return true
		}
	}
	return false
}
