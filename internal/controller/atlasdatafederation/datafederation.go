package atlasdatafederation

import (
	"errors"

	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/controller/workflow"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/translation/datafederation"
)

func (r *AtlasDataFederationReconciler) ensureDataFederation(ctx *workflow.Context, project *akov2.AtlasProject, dataFederation *akov2.AtlasDataFederation, federationService datafederation.DataFederationService) workflow.Result {
	projectID := project.ID()
	operatorSpec := &dataFederation.Spec

	akoDataFederation, err := datafederation.NewDataFederation(&dataFederation.Spec, projectID, nil)
	if err != nil {
		return workflow.Terminate(workflow.Internal, err)
	}

	atlasDataFederation, err := federationService.Get(ctx.Context, projectID, operatorSpec.Name)
	if err != nil {
		if !errors.Is(err, datafederation.ErrorNotFound) {
			return workflow.Terminate(workflow.Internal, err)
		}

		err = federationService.Create(ctx.Context, akoDataFederation)
		if err != nil {
			return workflow.Terminate(workflow.DataFederationNotCreatedInAtlas, err)
		}

		return workflow.InProgress(workflow.DataFederationCreating, "Data Federation is being created")
	}

	if akoDataFederation.SpecEqualsTo(atlasDataFederation) {
		return workflow.OK()
	}

	err = federationService.Update(ctx.Context, akoDataFederation)
	if err != nil {
		return workflow.Terminate(workflow.DataFederationNotUpdatedInAtlas, err)
	}

	return workflow.InProgress(workflow.DataFederationUpdating, "Data Federation is being updated")
}
