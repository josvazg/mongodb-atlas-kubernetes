package api

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type LocalRef string

// +k8s:deepcopy-gen=false

// CredentialsProvider gives access to custom local credentials
type CredentialsProvider interface {
	Credentials() *LocalObjectReference
}

// +k8s:deepcopy-gen=false

// ResourceWithCredentials is to be implemented by all CRDs using custom local credentials
type ResourceWithCredentials interface {
	CredentialsProvider
	GetName() string
	GetNamespace() string
}

// LocalCredentialHolder is to be embedded by Specs of CRDs using custom local credentials
type LocalCredentialHolder struct {
	ConnectionSecret *LocalObjectReference `json:"connectionSecret,omitempty"`
}

func (ch *LocalCredentialHolder) Credentials() *LocalObjectReference {
	return ch.ConnectionSecret
}

// +k8s:deepcopy-gen=false

// Reconciliable is implemented by CRD objects used by indexes to trigger reconciliations
type Reconciliable interface {
	ReconciliableRequests() []reconcile.Request
}

// ToRequest is a helper to turns CRD objects into reconcile requests
func ToRequest(obj client.Object) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		},
	}
}

// +k8s:deepcopy-gen=false

// ReconciliableList is a Reconciliable that is also a CRD list
type ReconciliableList interface {
	client.ObjectList
	Reconciliable
}
