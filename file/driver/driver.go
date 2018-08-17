package driver

import (
	"github.com/gluster/glusterd2/pkg/restclient"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/sirupsen/logrus"
)

const (
	driverName    = "org.gluster.glusterfs"
	vendorVersion = "1"
)

// Driver implements the following CSI interfaces:
//
//   csi.IdentityServer
//   csi.ControllerServer
//   csi.NodeServer
//
type Driver struct {
	csiDriver  *csicommon.CSIDriver
	client     *restclient.Client
	grpcserver csicommon.NonBlockingGRPCServer
	endpoint   string
	nodeID     string
	logger     *logrus.Logger
}

// NewDriver returns a CSI plugin to interact with Kubernetes
func NewDriver(nodeID, endpoint, glusterURL, username, secret string) *Driver {
	d := &Driver{}
	var err error
	d.nodeID = nodeID
	d.endpoint = endpoint
	d.client, err = restclient.New(glusterURL, username, secret, "", false)
	if err != nil {
		d.logger.Fatal("failed to initialize client")
	}
	d.logger = logrus.New()

	d.csiDriver = csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)

	d.logger.Info("Driver initalization successful")
	return d
}

func NewControllerServer(d *Driver) *ControllerServer {
	return &ControllerServer{
		Driver:                  d,
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d.csiDriver),
	}
}

func NewNodeServer(d *Driver) *NodeServer {
	return &NodeServer{
		Driver:            d,
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d.csiDriver),
	}
}

func NewidentityServer(d *Driver) *IdentityServer {
	return &IdentityServer{
		Driver:                d,
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d.csiDriver),
	}
}

//Run starts Non blocking GRPC server
func (d *Driver) Run() {
	d.grpcserver = csicommon.NewNonBlockingGRPCServer()
	d.grpcserver.Start(d.endpoint, NewidentityServer(d), NewControllerServer(d), NewNodeServer(d))
	d.grpcserver.Wait()
}

//Stop force stops running Non blocking GRPC server
func (d *Driver) Stop() {
	d.grpcserver.ForceStop()
}
