package interp

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

/*
variable-ref        =  "${" [namespace] variable-name "}"
namespace           =  identifier "." *sub-namespace
sub-namespace       =  variable-name "."
variable-name       =  num-variable / identifier
num-variable        =  1*DIGIT
*/
var variableRegexp = regexp.MustCompile(`\${(?:[a-zA-Z_][a-zA-Z0-9_]*\.(?:(?:[a-zA-Z_][a-zA-Z0-9_]*|[0-9]+)\.)*)?(?:[a-zA-Z_][a-zA-Z0-9_]*|[0-9]+)}`)

func usedVars(script *Script, s string) []string {
	if !script.RequiresExtension("variables") {
		return nil
	}

	variables := variableRegexp.FindAllString(s, -1)
	for i := range variables {
		// Cut ${} and case-fold.
		variables[i] = strings.ToLower(variables[i][2 : len(variables[i])-1])
	}

	return variables
}

func usedVarsAreValid(script *Script, s string) bool {
	for _, v := range usedVars(script, s) {
		matchNum, err := strconv.Atoi(v)
		if err == nil && matchNum >= 0 {
			continue
		}

		_, gettable := script.IsVarUsable(v)
		if !gettable {
			return false
		}
	}
	return true
}

func expandVarsList(d *RuntimeData, list []string) []string {
	if !d.Script.RequiresExtension("variables") {
		return list
	}

	listCpy := make([]string, len(list))
	for i, val := range list {
		listCpy[i] = expandVars(d, val)
	}
	return listCpy
}

func expandVars(d *RuntimeData, s string) string {
	if !d.Script.RequiresExtension("variables") {
		return s
	}

	expanded := variableRegexp.ReplaceAllStringFunc(s, func(match string) string {
		name := match[2 : len(match)-1]

		if matchNum, err := strconv.Atoi(name); err == nil && matchNum >= 0 {
			return d.MatchVariable(matchNum)
		}

		value, err := d.Var(name)
		if err != nil {
			panic("attempt to use an unusable variable: " + name)
		}
		return value
	})
	return expanded
}

type CmdSet struct {
	Name  string
	Value string

	ModifyValue func(string) string
}

func (c CmdSet) Execute(_ context.Context, d *RuntimeData) error {
	return d.SetVar(c.Name, c.ModifyValue(expandVars(d, c.Value)))
}

type TestString struct {
	Source     []string
	Key        []string
	Comparator Comparator
	Match      Match
}

func (t TestString) Check(_ context.Context, d *RuntimeData) (bool, error) {
	for _, source := range t.Source {
		source = expandVars(d, source)
		for _, key := range t.Key {
			key = expandVars(d, key)
			ok, matches, err := testString(t.Comparator, t.Match, source, key)
			if err != nil {
				return false, err
			}
			if ok {
				if t.Match == MatchMatches {
					d.MatchVariables = matches
				}
				return true, nil
			}
		}
	}
	return false, nil
}