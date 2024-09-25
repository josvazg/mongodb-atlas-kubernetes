// Sample pre-exiting code library, eg API SDK
package lib

import "context"

type Resource struct {
	ComplexSubtype ComplexSubtype
	Enabled        *bool
	Id             string
	OptionalRef    *OptionalRef
	SelectedOption *string
	Status         *string
}

type ComplexSubtype struct {
	Name    string
	Subtype string
}

type OptionalRef struct {
	Ref string
}

type API interface {
	Get(ctx context.Context, id string) (*Resource, error)
	Create(ctx context.Context, apiRes *Resource) (*Resource, error)
}

type DummyAPI struct {
}

func NewAPI() API {
	return &DummyAPI{}
}

func (d *DummyAPI) Get(ctx context.Context, id string) (*Resource, error) {
	panic("unimplemented")
}

func (d *DummyAPI) Create(ctx context.Context, apiRes *Resource) (*Resource, error) {
	panic("unimplemented")
}
