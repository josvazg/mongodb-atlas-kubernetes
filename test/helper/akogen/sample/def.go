// Sample internal types to define manually before code generation
package sample

// Resource is teh internal type
// +akogen:InternalType
// +akogen:ExternalSystem=Atlas
// +akogen:ExternalPackage=github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib
// +akogen:ExternalType=lib.Resource
// +akogen:ExternalInterface=lib.API
// +akogen:WrapperType:name="w",type="Wrapper"
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
