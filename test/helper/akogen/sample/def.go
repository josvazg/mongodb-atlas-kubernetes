// Sample internal types to define manually before code generation
package sample

// +akogen:ExternalSystem:Atlas
// +akogen:ExternalPackage:var=lib,path="github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"
// +akogen:ExternalType:var=res,type=*lib.Resource
// +akogen:ExternalAPI:var=api,type=lib.API
// +akogen:WrapperType:var="w",type="Wrapper"

// Resource is the internal type
// +akogen:InternalType:var=res,pointer=true
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
