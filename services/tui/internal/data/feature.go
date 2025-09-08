package data

import (
	"context"
	"io"
	"log"
	"time"

	tuiSvc "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/ui"
)

// DataManager manages the lifecycle of the TUI data stream.
// After proto update: client sends DataStreamReceived (acks/heartbeats),
// server sends DataStreamSend (snapshots). Heartbeat every 5s.

type DataManager struct {
	client tuiSvc.TUIServiceClient
	app    *ui.App
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDataManager(client tuiSvc.TUIServiceClient, app *ui.App) *DataManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DataManager{client: client, app: app, ctx: ctx, cancel: cancel}
}

// StartDataStream sets up bidirectional stream: we read server snapshots and send heartbeats.
func (dm *DataManager) StartDataStream() error {
	log.Println("[TUI] starting data stream...")
	stream, err := dm.client.SendDatastream(dm.ctx)
	if err != nil {
		return err
	}

	// Start receiver
	go dm.receiveSnapshots(stream)
	log.Println("[TUI] snapshot receiver started")
	// Start heartbeats
	go dm.sendHeartbeats(stream)
	return nil
}

// receiveSnapshots listens for DataStreamSend messages from server.
func (dm *DataManager) receiveSnapshots(stream tuiSvc.TUIService_SendDatastreamClient) {
	// (removed unused count variable)
	for {
		select {
		case <-dm.ctx.Done():
			return
		default:
			msg, err := stream.Recv()
			if err == nil {
				log.Printf("[TUI] snapshot received: hosts=%d logs_len=%d", len(msg.GetHostList().GetHosts()), len(msg.GetLogs()))
			}
			if err == io.EOF {
				log.Println("TUI stream closed by server")
				return
			}
			if err != nil {
				log.Printf("TUI stream receive error: %v", err)
				return
			}
			dm.processSnapshot(msg)
		}
	}
}

// sendHeartbeats periodically sends ack messages (DataStreamReceived) upstream.
func (dm *DataManager) sendHeartbeats(stream tuiSvc.TUIService_SendDatastreamClient) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-dm.ctx.Done():
			// attempt a final close-send
			_ = stream.CloseSend()
			return
		case <-ticker.C:
			if err := stream.Send(&tuiSvc.DataStreamReceived{Ack: "HEARTBEAT"}); err != nil {
				log.Printf("Heartbeat send failed: %v", err)
				return
			}
		}
	}
}

// processSnapshot updates UI based on snapshot
func (dm *DataManager) processSnapshot(data *tuiSvc.DataStreamSend) {
	if hl := data.GetHostList(); hl != nil {
		hosts := make([]ui.Host, 0, len(hl.GetHosts()))
		containersMap := make(map[string][]ui.Container, len(hl.GetHosts()))
		for _, hostInfo := range hl.GetHosts() {
			host := ConvertHostInfo(hostInfo)
			hosts = append(hosts, host)
			containers := make([]ui.Container, 0, len(hostInfo.GetContainers()))
			for _, cinfo := range hostInfo.GetContainers() {
				containers = append(containers, ConvertContainerInfo(cinfo))
			}
			containersMap[host.MACAddress] = containers
		}
		dm.app.QueueUpdateDraw(func() {
			dm.app.UpdateHosts(hosts)
			dm.app.UpdateContainersMap(containersMap)
		})
	}

	if ss := data.GetServicesStatus(); len(ss) > 0 {
		service := ConvertServicesStatus(ss)
		if hl := data.GetHostList(); hl != nil {
			service.TotalHosts = len(hl.GetHosts())
		}
		dm.app.QueueUpdateDraw(func() { dm.app.UpdateServices(service) })
	}

	if logs := data.GetLogs(); logs != "" {
		dm.app.QueueUpdateDraw(func() { dm.app.UpdateLogs(logs) })
	}

	if ct := data.GetCronTime(); ct > 0 {
		dm.app.QueueUpdateDraw(func() { dm.app.UpdateCronTime(ct) })
	}
}

// Stop terminates streaming.
func (dm *DataManager) Stop() { dm.cancel() }
