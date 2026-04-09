package palparser

import (
	"errors"

	"github.com/jaak-ai/passflow-api/src/usecase/pal/domain"
	"gopkg.in/yaml.v3"
)

var (
	// ErrInvalidYAML is returned when YAML parsing fails
	ErrInvalidYAML = errors.New("invalid YAML content")
	// ErrMissingAgentRoot is returned when agent root is missing from spec
	ErrMissingAgentRoot = errors.New("missing agent root in spec")
)

// ParseYAML parses YAML content into a PALSpec without validation
func ParseYAML(content []byte) (*domain.PALSpec, error) {
	var spec domain.PALSpec

	err := yaml.Unmarshal(content, &spec)
	if err != nil {
		return nil, ErrInvalidYAML
	}

	// Check if agent root is present
	if spec.Agent == nil {
		return nil, ErrMissingAgentRoot
	}

	return &spec, nil
}

// ParseAndValidate parses YAML content and validates the resulting PALSpec
func ParseAndValidate(content []byte) (*domain.PALSpec, error) {
	spec, err := ParseYAML(content)
	if err != nil {
		return nil, err
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	return spec, nil
}
