package atlasprivateendpoint

import (
	"context"
	"errors"
	"fmt"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api"
	akov2 "github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/controller/customresource"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/controller/workflow"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/indexer"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
)

// AtlasPrivateEndpointReconciler reconciles an AtlasPrivateEndpoint object
type AtlasPrivateEndpointReconciler struct {
	Client                   client.Client
	Log                      *zap.SugaredLogger
	Scheme                   *runtime.Scheme
	GlobalPredicates         []predicate.Predicate
	ObjectDeletionProtection bool
	// TODO: Decide
	// SubObjectDeletionProtection bool No need to add it right?
}

// NewAtlasPrivateEndpointReconciler creates a AtlasPrivateEndpointReconciler from the Manager
func NewAtlasPrivateEndpointReconciler(
	mgr manager.Manager,
	logger *zap.Logger,
	predicates []predicate.Predicate,
	deletionProtection bool,
) *AtlasPrivateEndpointReconciler {
	return &AtlasPrivateEndpointReconciler{
		Client:                   mgr.GetClient(),
		Log:                      logger.Named("controllers").Named("AtlasBackupCompliancePolicy").Sugar(),
		Scheme:                   mgr.GetScheme(),
		GlobalPredicates:         predicates,
		ObjectDeletionProtection: deletionProtection,
	}
}

func (r *AtlasPrivateEndpointReconciler) SetupWithManager(mgr ctrl.Manager, skipNameValidation bool) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("AtlasPrivateEndpointReconciler").
		For(&akov2.AtlasBackupCompliancePolicy{}, builder.WithPredicates(r.GlobalPredicates...)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(indexer.CredentialsIndexMapperFunc(
				indexer.AtlasPrivateEndpointCredentialsIndex,
				&akov2.AtlasPrivateEndpointList{},
				r.Client,
				r.Log,
			)),
			builder.WithPredicates(predicate.GenerationChangedPredicate{}),
		).
		WithOptions(controller.TypedOptions[reconcile.Request]{SkipNameValidation: pointer.MakePtr(skipNameValidation)}).
		Complete(r)
}

// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasprivateendpointreconciler,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasprivateendpointreconciler/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=atlasprivateendpointreconciler,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=atlasprivateendpointreconciler/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *AtlasPrivateEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.With("atlasprivateendpoint", req.NamespacedName)

	pe := &akov2.AtlasPrivateEndpoint{}
	result := customresource.PrepareResource(ctx, r.Client, req, pe, log)
	if !result.IsOk() {
		return result.ReconcileResult(), nil
	}

	if customresource.ReconciliationShouldBeSkipped(pe) {
		log.Infow(fmt.Sprintf("-> Skipping AtlasPrivateEndpoint reconciliation as annotation %s=%s",
			customresource.ReconciliationPolicyAnnotation, customresource.ReconciliationPolicySkip), "spec", pe.Spec)
		return workflow.OK().ReconcileResult(), nil
	}

	conditions := akov2.InitCondition(pe, api.FalseCondition(api.ReadyType))
	workflowCtx := workflow.NewContext(log, conditions, ctx)

	err := errors.New("not implemented")
	result = workflow.Terminate(workflow.AtlasFinalizerNotSet, err.Error())
	workflowCtx.SetConditionFromResult(api.PrivateEndpointReadyType, result)
	return result.ReconcileResult(), err
}
