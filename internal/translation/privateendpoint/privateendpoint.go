package privateendpoint

import (
	"context"
	"fmt"

	"go.mongodb.org/atlas-sdk/v20231115008/admin"
)

// PrivateServicesAPI is a Private Endpoint Service management API
type PrivateServicesAPI interface {
	ListServices(ctx context.Context, projectID, cloudProvider string) ([]EndpointService, error)
	CreateService(ctx context.Context, projectID, cloudProvider, region string) (*EndpointService, error)
	DeleteService(ctx context.Context, projectID, cloudProvider, endpointServiceID string) error
}

// PrivateConnectionAPI is a service to manage Private Endpoint connections
type PrivateConnectionAPI interface {
	Get(ctx context.Context, projectID, cloudProvider, endpointID, endpointServiceID string) (*PrivateEndpoint, error)
	Create(ctx context.Context, projectID, cloudProvider, endpointServiceID string, conn Connection) (*PrivateEndpoint, error)
	Delete(ctx context.Context, projectID, cloudProvider, endpointID, endpointServiceID string) error
}

type PrivateEndpointsAPI interface {
	PrivateServicesAPI
	PrivateConnectionAPI
}

type privateEndpointsAPI struct {
	peAPI admin.PrivateEndpointServicesApi
}

func NewPrivateEndpointsAPI(peAPI admin.PrivateEndpointServicesApi) PrivateEndpointsAPI {
	return &privateEndpointsAPI{peAPI: peAPI}
}

// Create implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) Create(ctx context.Context, projectID string, cloudProvider string, endpointServiceID string, conn Connection) (*PrivateEndpoint, error) {
	pe, _, err := p.peAPI.CreatePrivateEndpointWithParams(ctx, &admin.CreatePrivateEndpointApiParams{
		GroupId:               projectID,
		CloudProvider:         cloudProvider,
		EndpointServiceId:     endpointServiceID,
		CreateEndpointRequest: connToAtlas(&conn),
	}).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to Create PrivateEndpoint: %w", err)
	}
	return fromAtlas(pe), nil
}

// CreateService implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) CreateService(ctx context.Context, projectID string, cloudProvider, region string) (*EndpointService, error) {
	conn, _, err := p.peAPI.CreatePrivateEndpointService(ctx, projectID, &admin.CloudProviderEndpointServiceRequest{
		ProviderName: cloudProvider,
		Region:       region,
	}).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to Create EndpointService: %w", err)
	}
	return serviceFromAtlas(conn), nil
}

// Delete implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) Delete(ctx context.Context, projectID string, cloudProvider string, endpointID string, endpointServiceID string) error {
	_, _, err := p.peAPI.DeletePrivateEndpointWithParams(ctx, &admin.DeletePrivateEndpointApiParams{
		GroupId:           projectID,
		CloudProvider:     cloudProvider,
		EndpointId:        endpointID,
		EndpointServiceId: endpointServiceID,
	}).Execute()
	if err != nil {
		return fmt.Errorf("failed to Delete EndpointService: %w", err)
	}
	return nil
}

// DeleteService implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) DeleteService(ctx context.Context, projectID string, cloudProvider string, endpointServiceID string) error {
	_, _, err := p.peAPI.DeletePrivateEndpointServiceWithParams(ctx, &admin.DeletePrivateEndpointServiceApiParams{
		GroupId:           projectID,
		CloudProvider:     cloudProvider,
		EndpointServiceId: endpointServiceID,
	}).Execute()
	if err != nil {
		return fmt.Errorf("failed to Delete EndpointService: %w", err)
	}
	return nil
}

// Get implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) Get(ctx context.Context, projectID string, cloudProvider string, endpointID string, endpointServiceID string) (*PrivateEndpoint, error) {
	conn, _, err := p.peAPI.GetPrivateEndpointWithParams(ctx, &admin.GetPrivateEndpointApiParams{
		GroupId:           projectID,
		CloudProvider:     cloudProvider,
		EndpointId:        endpointID,
		EndpointServiceId: endpointServiceID,
	}).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to Get PrivateEndpointService: %w", err)
	}
	return fromAtlas(conn), nil
}

// ListServices implements PrivateEndpointsAPI.
func (p *privateEndpointsAPI) ListServices(ctx context.Context, projectID string, cloudProvider string) ([]EndpointService, error) {
	conns, _, err := p.peAPI.ListPrivateEndpointServices(ctx, projectID, cloudProvider).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to List PrivateEndpointServices: %w", err)
	}
	list := make([]EndpointService, 0, len(conns))
	for _, conn := range conns {
		list = append(list, *serviceFromAtlas(&conn))
	}
	return list, nil
}
