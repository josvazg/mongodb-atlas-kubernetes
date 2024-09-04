package status

import "github.com/mongodb/mongodb-atlas-kubernetes/v2/pkg/api"

// AtlasPrivateEndpointStatus defines the observed state of AtlasPrivateEndpoint
type AtlasPrivateEndpointStatus struct {
	api.Common `json:",inline"`
}

// +k8s:deepcopy-gen=false

type AtlasPrivateEndpointStatusOption func(s *AtlasPrivateEndpointStatus)
