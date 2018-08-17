package driver

import (
	"context"
	"fmt"
	"os"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume/util"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeServer struct {
	*csicommon.DefaultNodeServer
	*Driver
}

// NodeStageVolume mounts the volume to a staging path on the node.
func (d *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// NodeUnstageVolume unstages the volume from the staging path
func (d *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// NodePublishVolume mounts the volume mounted to the staging path to the target path
func (d *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	d.logger.Info("node publish volume called")

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target Path cannot be empty")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capability canont be empty")
	}

	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to created target path %s", err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Errorf(codes.Internal, "target path in not a moint point %s", err.Error())
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	mo := req.GetVolumeCapability().GetMount().GetMountFlags()
	if req.GetReadonly() {
		mo = append(mo, "ro")
	}

	s := req.GetVolumeAttributes()["glusterserver"]
	ep := req.GetVolumeAttributes()["glustervol"]
	source := fmt.Sprintf("%s:%s", s, ep)

	//TODO mounting with gluster backup server is pending
	mounter := mount.New("")
	err = mounter.Mount(source, targetPath, "glusterfs", mo)
	if err != nil {
		if os.IsPermission(err) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if strings.Contains(err.Error(), "invalid argument") {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	d.logger.Info("bind mounting the volume is finished")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (d *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target Path cannot be empty")
	}

	ll := d.logger.WithFields(logrus.Fields{
		"volume ID":    req.VolumeId,
		"Target Path":  req.TargetPath,
		"Request Type": "NodeUnpublishVolume",
	})
	ll.Info("node unpublish volume called")

	targetPath := req.GetTargetPath()
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Error(codes.NotFound, "Targetpath not found")
		}
		return nil, status.Error(codes.Internal, err.Error())

	}
	if notMnt {
		return nil, status.Error(codes.NotFound, "Volume not mounted")
	}

	err = util.UnmountPath(req.GetTargetPath(), mount.New(""))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ll.Info("unmounting volume is finished")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetId returns the unique id of the node.
func (d *NodeServer) NodeGetId(ctx context.Context, req *csi.NodeGetIdRequest) (*csi.NodeGetIdResponse, error) {
	d.logger.WithField("Request Type", "NodeGetId").Info("node get id requested")
	return &csi.NodeGetIdResponse{
		NodeId: d.Driver.nodeID,
	}, nil
}

//NodeGetInfo info
func (d *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	d.logger.WithField("Request Type", "NodeGetInfo").Info("node get info requested")
	return &csi.NodeGetInfoResponse{
		NodeId: d.Driver.nodeID,
	}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (d *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_UNKNOWN,
			},
		},
	}

	d.logger.WithFields(logrus.Fields{
		"Node capabilities": nscap,
		"Request Type":      "NodeGetCapabilities",
	}).Info("node get capabilities requested")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			nscap,
		},
	}, nil
}
