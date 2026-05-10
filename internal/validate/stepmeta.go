package validate

import (
	"fmt"
	"regexp"

	"github.com/justinush/maestro/internal/definition"
)

var annotationKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}$`)

const (
	maxStepDescriptionRunes = 8192
	maxStepLabels           = 64
	maxStepLabelRunes       = 128
	maxAnnotations          = 64
	maxAnnotationKeyRunes   = 64
	maxAnnotationValueRunes = 512
)

// validateStepMetadata enforces size/uniqueness and annotation key shape for step metadata.
func validateStepMetadata(def *definition.WorkflowDefinition) error {
	if def == nil {
		return nil
	}
	for i := range def.Steps {
		st := def.Steps[i]
		if err := validateOneStepMetadata(st.Description, st.Labels, st.Annotations); err != nil {
			return fmt.Errorf("step %q: %w", st.ID, err)
		}
	}
	return nil
}

func validateOneStepMetadata(description string, labels []string, annotations map[string]string) error {
	if len([]rune(description)) > maxStepDescriptionRunes {
		return fmt.Errorf("description exceeds %d characters", maxStepDescriptionRunes)
	}
	if len(labels) > maxStepLabels {
		return fmt.Errorf("labels: at most %d entries", maxStepLabels)
	}
	seen := make(map[string]struct{}, len(labels))
	for i, lab := range labels {
		if lab == "" {
			return fmt.Errorf("labels[%d]: empty string", i)
		}
		if len([]rune(lab)) > maxStepLabelRunes {
			return fmt.Errorf("labels[%d]: exceeds %d characters", i, maxStepLabelRunes)
		}
		if _, dup := seen[lab]; dup {
			return fmt.Errorf("labels[%d]: duplicate label %q", i, lab)
		}
		seen[lab] = struct{}{}
	}
	if len(annotations) > maxAnnotations {
		return fmt.Errorf("annotations: at most %d keys", maxAnnotations)
	}
	for k, v := range annotations {
		if len([]rune(k)) > maxAnnotationKeyRunes {
			return fmt.Errorf("annotations key %q exceeds %d characters", k, maxAnnotationKeyRunes)
		}
		if !annotationKeyPattern.MatchString(k) {
			return fmt.Errorf("annotations key %q: must match %s", k, annotationKeyPattern.String())
		}
		if len([]rune(v)) > maxAnnotationValueRunes {
			return fmt.Errorf("annotations[%q]: value exceeds %d characters", k, maxAnnotationValueRunes)
		}
	}
	return nil
}
