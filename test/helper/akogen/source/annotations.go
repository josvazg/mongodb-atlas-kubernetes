package source

import (
	"fmt"
	"go/ast"
	"strings"
)

type AnnotationType int

const (
	NoValue AnnotationType = iota
	SimpleValue
	ArgsValues
)

// GenAnnotation represent any annotation in code with the following format
// {generator-name}:{Name}:{Args}
// Args is a comma separated list of key=value pairs. Values might be in quotes.
//
// Eg. `+akogen:ExternalAPI:var=api,type="lib.API"`
type GenAnnotation struct {
	Raw   string
	Name  string
	Type  AnnotationType
	Args  map[string]string
	Value string
}

func (sg *SourceGen) GenAnnotationsFor(generatorName string) ([]GenAnnotation, error) {
	pattern := generatorName
	if pattern[0] != '+' {
		pattern = "+" + pattern
	}

	annotations := []GenAnnotation{}
	for _, files := range sg.pkgs {
		for _, f := range files {
			var err error
			annotations, err = genAnnotationsFor(annotations, pattern, f)
			if err != nil {
				return nil, fmt.Errorf("failed to extract annotations on file %s: %w", f.Name, err)
			}
		}
	}
	return annotations, nil
}

func genAnnotationsFor(annotations []GenAnnotation, pattern string, f *ast.File) ([]GenAnnotation, error) {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if !strings.Contains(c.Text, pattern) {
				continue
			}
			payload, err := parseAnnotationValue(c.Text)
			if err != nil {
				return nil, fmt.Errorf("failed to parse annotation value: %w", err)
			}
			ga, err := buildGenAnnotationValue(c.Text, payload)
			if err != nil {
				return nil, fmt.Errorf("failed to parse annotation: %w", err)
			}
			annotations = append(annotations, ga)
		}
	}
	return annotations, nil
}

func buildGenAnnotationValue(raw, payload string) (GenAnnotation, error) {
	parts := strings.Split(payload, ":")
	if len(parts) < 1 || len(parts) > 2 {
		return GenAnnotation{}, fmt.Errorf("failed to parse annotation name for %q", payload)
	}
	if len(parts) == 1 || len(parts) == 2 && strings.TrimSpace(parts[1]) == "" {
		return GenAnnotation{Raw: raw, Name: parts[0], Type: NoValue}, nil
	}
	name := strings.TrimSpace(parts[0])
	args, err := parseAnnotationCSVValue(payload)
	if err == nil {
		return GenAnnotation{Raw: raw, Name: name, Type: ArgsValues, Args: args}, nil
	}
	value, err := parseAnnotationValue(payload)
	if err == nil {
		return GenAnnotation{Raw: raw, Name: name, Type: SimpleValue, Value: value}, nil
	}
	return GenAnnotation{}, fmt.Errorf("failed to parse annotation name and value for %q", payload)
}

func parseAnnotationValue(s string) (string, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to extract annotation value from %s", s)
	}
	return parts[1], nil
}

func parseAnnotationCSVValue(s string) (map[string]string, error) {
	value, err := parseAnnotationValue(s)
	if err != nil {
		return nil, err
	}
	return parseCSVMap(value)
}

func parseCSVMap(value string) (map[string]string, error) {
	values := map[string]string{}
	tuples := strings.Split(value, ",")
	for _, tuple := range tuples {
		k, v, err := parseAssignment(tuple)
		if err != nil {
			return nil, fmt.Errorf("failed to parse annotation CSV values: %w", err)
		}
		values[k] = v
	}
	return values, nil
}

func parseAssignment(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("failed to extract assignment value from %q", s)
	}
	return parts[0], strings.Trim(parts[1], " \""), nil
}
