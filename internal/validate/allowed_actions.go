package validate

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var (
	builtinActionTypes = []string{"stub", "http"}

	// actionTypeNamePattern matches kebab-case identifiers.
	actionTypeNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
)

func (o Options) withNormalizedAllowedActionTypes() (Options, error) {
	if len(o.AllowedActionTypes) == 0 {
		return o, nil
	}
	allowed, err := normalizeAllowedActionTypes(o.AllowedActionTypes)
	if err != nil {
		return Options{}, err
	}
	o.AllowedActionTypes = allowed
	return o, nil
}

func normalizeAllowedActionTypes(types []string) ([]string, error) {
	if len(types) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(types))
	out := make([]string, 0, len(types))

	for i, raw := range types {
		t := strings.TrimSpace(raw)
		if t == "" {
			return nil, fmt.Errorf("allowed action types[%d]: empty", i)
		}
		if slices.Contains(builtinActionTypes, t) {
			return nil, fmt.Errorf("allowed action types[%d]: %q is built-in; omit it from AllowedActionTypes", i, t)
		}
		if !actionTypeNamePattern.MatchString(t) {
			return nil, fmt.Errorf("allowed action types[%d]: %q must match %s", i, t, actionTypeNamePattern.String())
		}
		if _, ok := seen[t]; ok {
			return nil, fmt.Errorf("allowed action types: duplicate %q", t)
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}

	slices.Sort(out)
	return out, nil
}

func extendActionTypeEnum(schemaDoc map[string]any, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}

	defs, ok := schemaDoc["$defs"].(map[string]any)
	if !ok {
		return fmt.Errorf("schema: missing $defs")
	}
	action, ok := defs["action"].(map[string]any)
	if !ok {
		return fmt.Errorf("schema: missing $defs.action")
	}
	props, ok := action["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("schema: missing $defs.action.properties")
	}
	typeProp, ok := props["type"].(map[string]any)
	if !ok {
		return fmt.Errorf("schema: missing $defs.action.properties.type")
	}

	enumAny, ok := typeProp["enum"].([]any)
	if !ok {
		return fmt.Errorf("schema: action.type enum is not an array")
	}

	seen := make(map[string]struct{}, len(enumAny)+len(allowed))
	for _, v := range enumAny {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("schema: action.type enum contains non-string")
		}
		seen[s] = struct{}{}
	}
	for _, t := range allowed {
		if _, ok := seen[t]; ok {
			continue
		}
		enumAny = append(enumAny, t)
		seen[t] = struct{}{}
	}
	typeProp["enum"] = enumAny
	return nil
}
