// Sample internal types to define manually before code generation
package sample

type Resource struct {
	ComplexSubtype ComplexSubtype
	Enabled        bool
	ID             string
	OptionalRef    *OptionalRef
	SelectedOption OptionType
	Status         string
}

type ComplexSubtype struct {
	Name    string
	Subtype Subtype
}

type OptionalRef struct {
	Ref string
}

type Subtype string

type OptionType string
