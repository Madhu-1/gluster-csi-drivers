package driver

import (
	"context"
	"fmt"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/gluster/glusterd2/pkg/api"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
	TB
)

const (
	defaultVolumeSizeInMB = 20 * MB
	originKey             = "Created-BY"
	originValue           = "Gluster-CSI-Diver"
)

type ControllerServer struct {
	*csicommon.DefaultControllerServer
	*Driver
}

//CreateVolume calls glusterd2 to create and start volume
func (d *ControllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "volume name cannot be emty")
	}
	d.logger.Info("creating volume with name ", req.Name)
	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume capabilities is required")
	}

	volumeName := req.Name
	//get volume and check volume is already created
	//this is required during volume creation if CSI driver goes down
	//it should not create extra volumes
	volumes, err := d.client.Volumes("")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list volumes %s", err.Error())
	}
	// Volume Size - Default is 20 MB
	volSizeBytes := defaultVolumeSizeInMB
	if req.GetCapacityRange() != nil {
		volSizeBytes = int(req.GetCapacityRange().GetRequiredBytes())
	}
	volSizeMB := int((volSizeBytes + MB - 1) / MB)

	for _, vol := range volumes {
		if vol.Name == volumeName {
			// check if it was created by the CSI driver
			if createdBy, found := vol.Metadata[originKey]; found {
				if createdBy != originValue {
					return nil, status.Errorf(codes.Internal, "volume %s (%s) was not created by Guster CSI driver",
						vol.Name, vol.Metadata)
				}
			} else {
				return nil, status.Errorf(codes.Internal, "volume %s (%s) was not created by Guster CSI driver",
					vol.Name, vol.Metadata)
			}
			v, e := d.client.VolumeStatus(vol.ID.String())
			if e != nil {
				return nil, status.Errorf(codes.Internal, "failed to get volume status %s", e.Error())
			}
			if int(v.Size.Capacity) != volSizeBytes {
				return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("volume already exits with diferent size: %d", v.Size.Capacity))
			}

			d.logger.Info("volume already exists")
			glusterServer, peeraddresses, err := d.getpeeraddress()

			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("error in fecthing peer details %s", err.Error()))
			}

			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					Id:            vol.Name,
					CapacityBytes: int64(volSizeBytes),
					Attributes: map[string]string{
						"glustervol":        volumeName,
						"glusterserver":     glusterServer,
						"glusterbkpservers": strings.Join(peeraddresses, ":"),
					},
				},
			}, nil

		}
	}

	ll := d.logger.WithFields(logrus.Fields{
		"name":       volumeName,
		"size in MB": volSizeMB,
	})

	m := make(map[string]string)
	m[originKey] = originValue
	volumeReq := api.VolCreateReq{
		Name:     volumeName,
		Metadata: m,
		Size:     uint64(volSizeMB),
	}

	ll.WithField("volume request", volumeReq).Debug("creating volume")

	volume, err := d.client.VolumeCreate(volumeReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error creating volume %s", err.Error()))
	}

	err = d.client.VolumeStart(volumeName, true)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in starting volume %s", err.Error()))
	}

	glusterVol := volume.Name
	glusterServer, peeraddresses, err := d.getpeeraddress()

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in fecthing peer details %s", err.Error()))
	}

	resp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volume.Name,
			CapacityBytes: int64(volSizeBytes),
			Attributes: map[string]string{
				"glustervol":        glusterVol,
				"glusterserver":     glusterServer,
				"glusterbkpservers": strings.Join(peeraddresses, ":"),
			},
		},
	}

	ll.WithField("volume response", resp).Info("volume created")
	return resp, nil
}

func (d *ControllerServer) getpeeraddress() (string, []string, error) {

	peers, err := d.client.Peers()
	if err != nil {
		return "", nil, err
	}
	glusterServer := ""
	peeraddresses := []string{}

	for i, p := range peers {
		if i == 0 {
			for _, a := range p.PeerAddresses {
				if !strings.Contains(a, "127.0.0.1") {
					ip := strings.Split(a, ":")
					glusterServer = ip[0]
				}
			}
			continue
		}
		for _, a := range p.PeerAddresses {
			if !strings.Contains(a, "127.0.0.1") {
				ip := strings.Split(a, ":")
				peeraddresses = append(peeraddresses, ip[0])
			}
		}
	}
	return glusterServer, peeraddresses, err
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (d *ControllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "To delete volume Volume ID must be provided")
	}

	ll := d.logger.WithFields(logrus.Fields{
		"volume ID ":    req.VolumeId,
		"Request Type ": "DELETE VOLUME",
	})
	ll.Info("delete volume called")

	err := d.client.VolumeStop(req.VolumeId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error starting volume %s", err.Error())
	}

	err = d.client.VolumeDelete(req.VolumeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error deleting volume %s", err.Error())
	}
	ll.WithField("volume ID", req.VolumeId).Info("volume is deleted")
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *ControllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

//ControllerUnpublishVolume deattaches the given volume from the node
func (d *ControllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ValidateVolumeCapabilities checks whether the volume capabilities requested
// are supported.
func (d *ControllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID cannot be empty")
	}

	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities is required field")
	}

	var vcaps []*csi.VolumeCapability_AccessMode
	for _, mode := range []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	} {
		vcaps = append(vcaps, &csi.VolumeCapability_AccessMode{Mode: mode})
	}

	ll := d.logger.WithFields(logrus.Fields{
		"Volume ID":              req.VolumeId,
		"volume capabilities":    req.VolumeCapabilities,
		"supported capabilities": vcaps,
		"Request Type":           "ValidateVolumeCapabilities",
	})
	ll.Info("validate volume capabilities called")

	hasSupport := func(mode csi.VolumeCapability_AccessMode_Mode) bool {
		for _, m := range vcaps {
			if mode == m.Mode {
				return true
			}
		}
		return false
	}

	resp := &csi.ValidateVolumeCapabilitiesResponse{
		Supported: false,
	}

	for _, cap := range req.VolumeCapabilities {
		if hasSupport(cap.AccessMode.Mode) {
			resp.Supported = true
		} else {
			resp.Supported = false
		}
	}
	ll.WithField("response", resp).Info("volume supported capabilities")
	return resp, nil
}

// ListVolumes returns a list of volumes
func (d *ControllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	volumes, err := d.client.Volumes("")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list volumes %s", err.Error())
	}
	var entries []*csi.ListVolumesResponse_Entry
	for _, vol := range volumes {
		v, e := d.client.VolumeStatus(vol.Name)
		if e != nil {
			return nil, status.Errorf(codes.Internal, "failed to get volume status %s", err.Error())
		}
		entries = append(entries, &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				Id:            vol.ID.String(),
				CapacityBytes: int64(v.Size.Capacity * MB),
			},
		})
	}

	resp := &csi.ListVolumesResponse{
		Entries: entries,
	}

	return resp, nil

}

// GetCapacity returns the capacity of the storage pool
func (d *ControllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	d.logger.WithFields(logrus.Fields{
		"request":      req.Parameters,
		"Request Type": "Get Capacity",
	}).Warn("get capacity is not implemented")
	return nil, status.Error(codes.Unimplemented, "get capacity not implemented")
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (d *ControllerServer) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var caps []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	return resp, nil
}

//CreateSnapshot as
func (d *ControllerServer) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

//DeleteSnapshot as
func (d *ControllerServer) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

//ListSnapshots as
func (d *ControllerServer) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
