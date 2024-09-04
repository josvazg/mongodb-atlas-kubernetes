/*
Copyright 2024 MongoDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1/common"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1/provider"
	"github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api/v1/status"
)

// Important:
// The procedure working with this file:
// 1. Edit the file
// 1. Run "make generate" to regenerate code
// 2. Run "make manifests" to regenerate the CRD

func init() {
	SchemeBuilder.Register(&AtlasPrivateEndpoint{}, &AtlasPrivateEndpointList{})
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Atlas ID",type=string,JSONPath=`.status.id`
// +kubebuilder:resource:categories=atlas,shortName=ape

// AtlasPrivateEndpoint is the Schema for the Atlas Private Endpoint API
type AtlasPrivateEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AtlasPrivateEndpointSpec          `json:"spec,omitempty"`
	Status status.AtlasPrivateEndpointStatus `json:"status,omitempty"`
}

// AtlasPrivateEndpointSpec contains the desired to configuration for an Atlas Private Endpoint
type AtlasPrivateEndpointSpec struct {
	api.LocalCredentialHolder `json:",inline"`

	AtlasPrivateEndpointConfig `json:",inline"`

	// Project is a reference to AtlasProject resource the user belongs to
	Project common.ResourceRefNamespaced `json:"projectRef"`
}

// AtlasPrivateEndpointConfig contains the pure Atlas side settings, without any Kubernetes related field
type AtlasPrivateEndpointConfig struct {
	// Cloud provider for which you want to retrieve a private endpoint service. Atlas accepts AWS or AZURE.
	// +kubebuilder:validation:Enum=AWS;GCP;AZURE;TENANT
	Provider provider.ProviderName `json:"provider"`
	// Cloud provider region for which you want to create the private endpoint service.
	Region string `json:"region"`
	// Unique identifier of the private endpoint you created in your AWS VPC or Azure Vnet.
	// +optional
	ID string `json:"id,omitempty"`
	// Private IP address of the private endpoint network interface you created in your Azure VNet.
	// +optional
	IP string `json:"ip,omitempty"`
	// Unique identifier of the Google Cloud project in which you created your endpoints.
	// +optional
	GCPProjectID string `json:"gcpProjectId,omitempty"`
	// Unique identifier of the endpoint group. The endpoint group encompasses all of the endpoints that you created in Google Cloud.
	// +optional
	EndpointGroupName string `json:"endpointGroupName,omitempty"`
	// Collection of individual private endpoints that comprise your endpoint group.
	// +optional
	Endpoints GCPEndpoints `json:"endpoints,omitempty"`
}

// GetStatus implements api.Reader
func (p *AtlasPrivateEndpoint) GetStatus() api.Status {
	return p.Status
}

// UpdateStatus implements api.Writer
func (p *AtlasPrivateEndpoint) UpdateStatus(conditions []api.Condition, options ...api.Option) {
	p.Status.Conditions = conditions
	p.Status.ObservedGeneration = p.ObjectMeta.Generation

	for _, o := range options {
		// This will fail if the Option passed is incorrect - which is expected
		v := o.(status.AtlasPrivateEndpointStatusOption)
		v(&p.Status)
	}
}

// Credentials implements api.Credentials
func (p AtlasPrivateEndpoint) Credentials() *api.LocalObjectReference {
	return p.Spec.Credentials()
}

// +kubebuilder:object:root=true

// AtlasPrivateEndpointList contains a list of AtlasPrivateEndpoint
type AtlasPrivateEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AtlasPrivateEndpoint `json:"items"`
}

// ReconciliableRequests implements ReconciliableList for Private Endpoints
func (list *AtlasPrivateEndpointList) ReconciliableRequests() []reconcile.Request {
	requests := make([]reconcile.Request, 0, len(list.Items))
	for _, item := range list.Items {
		requests = append(requests, api.ToRequest(&item))
	}
	return requests
}
