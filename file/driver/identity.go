package driver

import (
	"context"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/sirupsen/logrus"
)

type IdentityServer struct {
	*csicommon.DefaultIdentityServer
	*Driver
}

// GetPluginInfo returns metadata of the plugin
func (d *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	resp := &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: vendorVersion,
	}
	d.logger.WithFields(logrus.Fields{
		"response":     resp,
		"Request Type": "GetPluginInfo",
	}).Info("get plugin info requested")
	return resp, nil
}

// GetPluginCapabilities returns available capabilities of the plugin
func (d *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}

	d.logger.WithFields(logrus.Fields{
		"Response ":    resp,
		"Request Type": "GetPluginCapabilities",
	}).Info("get plugin capabitilies requested")
	return resp, nil
}

// Probe returns the health and readiness of the plugin
func (d *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	d.logger.WithField("Request Type", "probe").Info("probe requested")
	return &csi.ProbeResponse{}, nil
}
