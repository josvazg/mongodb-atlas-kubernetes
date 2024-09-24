// Generated by AKOGen code Generator - do not edit

package sample

import (
	"context"
	pointer "github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/test/helper/akogen/lib"
)

type wrapper struct {
	api lib.API
}

func (w *wrapper) Get(ctx context.Context, id string) (*Resource, error) {
	apiRes, err := w.api.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromAtlas(apiRes), nil
}

func (w *wrapper) Create(ctx context.Context, res *Resource) (*Resource, error) {
	apiRes, err := w.api.Create(ctx, toAtlas(res))
	if err != nil {
		return nil, err
	}
	return fromAtlas(apiRes), nil
}

func toAtlas(res *Resource) *lib.Resource {
	if res == nil {
		return nil
	}
	return &lib.Resource{
		ComplexSubtype: complexSubtypeToAtlas(res.ComplexSubtype),
		Enabled:        pointer.MakePtr(res.Enabled),
		Id:             res.ID,
		OptionalRef:    optionalRefToAtlas(res.OptionalRef),
		SelectedOption: pointer.MakePtr(string(res.SelectedOption)),
		Status:         pointer.MakePtr(res.Status),
	}
}

func fromAtlas(apiRes *lib.Resource) *Resource {
	if apiRes == nil {
		return nil
	}
	return &Resource{
		ComplexSubtype: complexSubtypeFromAtlas(apiRes.ComplexSubtype),
		Enabled:        pointer.GetOrDefault(apiRes.Enabled, false),
		ID:             apiRes.Id,
		OptionalRef:    optionalRefFromAtlas(apiRes.OptionalRef),
		SelectedOption: OptionType(pointer.GetOrDefault(apiRes.SelectedOption, "")),
		Status:         pointer.GetOrDefault(apiRes.Status, ""),
	}
}

func complexSubtypeToAtlas(cst ComplexSubtype) lib.ComplexSubtype {
	return lib.ComplexSubtype{
		Name:    cst.Name,
		Subtype: string(cst.Subtype),
	}
}

func complexSubtypeFromAtlas(apiCst lib.ComplexSubtype) ComplexSubtype {
	return ComplexSubtype{
		Name:    apiCst.Name,
		Subtype: Subtype(apiCst.Subtype),
	}
}

func optionalRefToAtlas(or *OptionalRef) *lib.OptionalRef {
	if or == nil {
		return nil
	}
	return &lib.OptionalRef{Ref: or.Ref}
}

func optionalRefFromAtlas(apiOr *lib.OptionalRef) *OptionalRef {
	if apiOr == nil {
		return nil
	}
	return &OptionalRef{Ref: apiOr.Ref}
}

// Generated by AKOGen code Generator - do not edit
