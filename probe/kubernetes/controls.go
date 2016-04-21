package kubernetes

import (
	"io"
	"io/ioutil"

	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/controls"
	"github.com/weaveworks/scope/report"
)

// Control IDs used by the kubernetes integration.
const (
	GetLogs       = "kubernetes_get_logs"
	DeletePod     = "kubernetes_delete_pod"
	DeleteService = "kubernetes_delete_service"
)

// GetLogs is the control to get the logs for a kubernetes pod
func (r *Reporter) GetLogs(req xfer.Request, namespaceID, podID string) xfer.Response {
	readCloser, err := r.client.GetLogs(namespaceID, podID)
	if err != nil {
		return xfer.ResponseError(err)
	}

	readWriter := struct {
		io.Reader
		io.Writer
	}{
		readCloser,
		ioutil.Discard,
	}
	id, pipe, err := controls.NewPipeFromEnds(nil, readWriter, r.pipes, req.AppID)
	if err != nil {
		return xfer.ResponseError(err)
	}
	pipe.OnClose(func() {
		readCloser.Close()
	})
	return xfer.Response{
		Pipe: id,
	}
}

func (r *Reporter) deletePod(_ xfer.Request, namespaceID, podID string) xfer.Response {
	return xfer.ResponseError(r.client.DeletePod(namespaceID, podID))
}

func (r *Reporter) deleteService(_ xfer.Request, namespaceID, serviceID string) xfer.Response {
	return xfer.ResponseError(r.client.DeleteService(namespaceID, serviceID))
}

func capturePod(f func(xfer.Request, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		namespaceID, podID, ok := report.ParsePodNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		return f(req, namespaceID, podID)
	}
}

func captureService(f func(xfer.Request, string, string) xfer.Response) func(xfer.Request) xfer.Response {
	return func(req xfer.Request) xfer.Response {
		namespaceID, serviceID, ok := report.ParseServiceNodeID(req.NodeID)
		if !ok {
			return xfer.ResponseErrorf("Invalid ID: %s", req.NodeID)
		}
		return f(req, namespaceID, serviceID)
	}
}

func (r *Reporter) registerControls() {
	controls.Register(GetLogs, capturePod(r.GetLogs))
	controls.Register(DeletePod, capturePod(r.deletePod))
	controls.Register(DeleteService, captureService(r.deleteService))
}

func (r *Reporter) deregisterControls() {
	controls.Rm(GetLogs)
	controls.Rm(DeletePod)
	controls.Rm(DeleteService)
}
