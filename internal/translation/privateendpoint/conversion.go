package privateendpoint

import (
	"go.mongodb.org/atlas-sdk/v20231115008/admin"

	"github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/pointer"
)

type Connection struct {
}

type EndpointService struct {
}

type PrivateEndpointConfig struct {
	// Cloud provider for which you want to retrieve a private endpoint service. Atlas accepts AWS or AZURE.
	// +kubebuilder:validation:Enum=AWS;GCP;AZURE;TENANT
	Provider ProviderName `json:"provider"`
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

type PrivateEndpointNotSureWhatThisIs struct {
	// Unique identifier for AWS or AZURE Private Link Connection.
	ID string `json:"id,omitempty"`
	// Cloud provider for which you want to retrieve a private endpoint service. Atlas accepts AWS or AZURE.
	Provider ProviderName `json:"provider"`
	// Cloud provider region for which you want to create the private endpoint service.
	Region string `json:"region"`
	// Name of the AWS or Azure Private Link Service that Atlas manages.
	ServiceName string `json:"serviceName,omitempty"`
	// Unique identifier of the Azure Private Link Service (for AWS the same as ID).
	ServiceResourceID string `json:"serviceResourceId,omitempty"`
	// Unique identifier of the AWS or Azure Private Link Interface Endpoint.
	InterfaceEndpointID string `json:"interfaceEndpointId,omitempty"`
	// Unique alphanumeric and special character strings that identify the service attachments associated with the GCP Private Service Connect endpoint service.
	ServiceAttachmentNames []string `json:"serviceAttachmentNames,omitempty"`
	// Collection of individual GCP private endpoints that comprise your network endpoint group.
	Endpoints []GCPEndpoint `json:"endpoints,omitempty"`
}

type ProviderName string

type GCPEndpoints []GCPEndpoint

type GCPEndpointConfig struct {
	// Forwarding rule that corresponds to the endpoint you created in Google Cloud.
	EndpointName string `json:"endpointName,omitempty"`
	// Private IP address of the endpoint you created in Google Cloud.
	IPAddress string `json:"ipAddress,omitempty"`
}

type GCPEndpoint struct {
	GCPEndpointConfig
	Status string
}

func serviceFromAtlas(pes *admin.EndpointService) *EndpointService {
	panic("unimplemented") //return EndpointService{}
}

type PrivateEndpointStatus struct {
	DeleteRequested               bool
	ErrorMessage                  string
	ConnectionStatus              string
	PrivateEndpointConnectionName string
	PrivateEndpointIPAddress      string
	Status                        string
}

type PrivateEndpoint struct {
	PrivateEndpointConfig
	PrivateEndpointStatus
}

/*
func toAtlas(pe *PrivateEndpoint) *admin.PrivateLinkEndpoint {
	return &admin.PrivateLinkEndpoint{
		// Config
		CloudProvider:            string(pe.Provider),
		InterfaceEndpointId:      pointer.MakePtr(pe.ID),
		PrivateEndpointIPAddress: pointer.MakePtr(pe.IP),
		EndpointGroupName:        pointer.MakePtr(pe.EndpointGroupName),
		Endpoints:                toAtlasGCEEndpoints(pe.Endpoints),
		// Status
		DeleteRequested:               pointer.MakePtr(pe.DeleteRequested),
		ErrorMessage:                  pointer.MakePtr(pe.ErrorMessage),
		ConnectionStatus:              pointer.MakePtr(pe.ConnectionStatus),
		PrivateEndpointConnectionName: pointer.MakePtr(pe.PrivateEndpointConnectionName),
		PrivateEndpointResourceId:     pointer.MakePtr(pe.PrivateEndpointIPAddress),
		Status:                        pointer.MakePtr(pe.Status),
	}
}

func toAtlasGCEEndpoints(geps GCPEndpoints) *[]admin.GCPConsumerForwardingRule {
	if geps == nil {
		return nil
	}
	rules := make([]admin.GCPConsumerForwardingRule, 0, len(geps))
	for _, gep := range geps {
		rule := admin.GCPConsumerForwardingRule{
			EndpointName: &gep.EndpointName,
			IpAddress:    &gep.IPAddress,
			Status:       new(string),
		}
		rules = append(rules, rule)
	}
	return &rules
}
*/

func fromAtlas(ple *admin.PrivateLinkEndpoint) *PrivateEndpoint {
	return &PrivateEndpoint{
		PrivateEndpointConfig: PrivateEndpointConfig{
			Provider:          ProviderName(ple.CloudProvider),
			ID:                pointer.GetOrDefault(ple.InterfaceEndpointId, ""),
			IP:                pointer.GetOrDefault(ple.PrivateEndpointIPAddress, ""),
			EndpointGroupName: pointer.GetOrDefault(ple.EndpointGroupName, ""),
			Endpoints:         fromAtlasGCEEndpoints(ple.Endpoints),
		},
		PrivateEndpointStatus: PrivateEndpointStatus{
			DeleteRequested:               pointer.GetOrDefault(ple.DeleteRequested, false),
			ErrorMessage:                  pointer.GetOrDefault(ple.ErrorMessage, ""),
			ConnectionStatus:              pointer.GetOrDefault(ple.ConnectionStatus, ""),
			PrivateEndpointConnectionName: pointer.GetOrDefault(ple.PrivateEndpointConnectionName, ""),
			PrivateEndpointIPAddress:      pointer.GetOrDefault(ple.PrivateEndpointIPAddress, ""),
			Status:                        pointer.GetOrDefault(ple.Status, ""),
		},
	}
}

func fromAtlasGCEEndpoints(gfr *[]admin.GCPConsumerForwardingRule) []GCPEndpoint {
	if gfr == nil {
		return nil
	}
	rules := make([]GCPEndpoint, 0, len(*gfr))
	for _, fr := range *gfr {
		rule := GCPEndpoint{
			GCPEndpointConfig: GCPEndpointConfig{
				EndpointName: pointer.GetOrDefault(fr.EndpointName, ""),
				IPAddress:    pointer.GetOrDefault(fr.IpAddress, ""),
			},
			Status: pointer.GetOrDefault(fr.Status, ""),
		}
		rules = append(rules, rule)
	}
	return rules
}

func connToAtlas(conn *Connection) *admin.CreateEndpointRequest {
	panic("unimplemented") // return &admin.CreateEndpointRequest{}
}
