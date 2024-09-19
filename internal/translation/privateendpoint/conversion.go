package privateendpoint

import (
	"go.mongodb.org/atlas-sdk/v20231115008/admin"
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

type PrivateEndpoint struct {
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

type GCPEndpoint struct {
	// Forwarding rule that corresponds to the endpoint you created in Google Cloud.
	EndpointName string `json:"endpointName,omitempty"`
	// Private IP address of the endpoint you created in Google Cloud.
	IPAddress string `json:"ipAddress,omitempty"`
}

func fromAtlas(pes *admin.PrivateLinkEndpoint) *PrivateEndpoint {
	panic("unimplemented") //return &PrivateEndpointService{}
}

func serviceFromAtlas(pes *admin.EndpointService) *EndpointService {
	panic("unimplemented") //return EndpointService{}
}

func toAtlas(pe *PrivateEndpointConfig) *admin.PrivateLinkEndpoint {
	return &admin.PrivateLinkEndpoint{
		CloudProvider:                 string(pe.Provider),
		DeleteRequested:               new(bool),
		ErrorMessage:                  new(string),
		ConnectionStatus:              new(string),
		InterfaceEndpointId:           &pe.ID,
		PrivateEndpointConnectionName: new(string),
		PrivateEndpointIPAddress:      &pe.IP,
		PrivateEndpointResourceId:     new(string),
		Status:                        new(string),
		EndpointGroupName:             &pe.EndpointGroupName,
		Endpoints:                     toAtlasGCEEndpoints(pe.Endpoints),
	}
}

func connToAtlas(conn *Connection) *admin.CreateEndpointRequest {
	panic("unimplemented") // return &admin.CreateEndpointRequest{}
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

type atlasPE struct {
	admin.EndpointService
}
