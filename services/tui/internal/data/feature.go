package data

import (
	"context"
	"io"
	"log"
	"time"

	orchestrator "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/ui"
)

type DataManager struct {
	client orchestrator.TUIServiceClient
	app    *ui.App
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDataManager(client orchestrator.TUIServiceClient, app *ui.App) *DataManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DataManager{
		client: client,
		app:    app,
		ctx:    ctx,
		cancel: cancel,
	}
}

// StartDataStream initiates the gRPC streaming connection
func (dm *DataManager) StartDataStream() error {
	stream, err := dm.client.SendDatastream(dm.ctx)
	if err != nil {
		return err
	}

	// Send initial acknowledgment
	err = stream.Send(&orchestrator.DataStreamreq{
		Ack: "TUI_READY",
	})
	if err != nil {
		return err
	}

	// Start receiving data in goroutine
	go dm.handleDataStream(stream)

	// Send periodic heartbeats
	go dm.sendHeartbeats(stream)

	return nil
}

// handleDataStream processes incoming data from the server
func (dm *DataManager) handleDataStream(stream orchestrator.TUIService_SendDatastreamClient) {
	for {
		select {
		case <-dm.ctx.Done():
			return
		default:
			data, err := stream.Recv()
			if err == io.EOF {
				log.Println("Stream ended by server")
				return
			}
			if err != nil {
				log.Printf("Error receiving  %v", err)
				continue
			}

			// Process received data
			dm.processDataStream(data)
		}
	}
}

// processDataStream updates UI components with new data
func (dm *DataManager) processDataStream(data *orchestrator.DataStreamSend) {
	// Update hosts data
	if len(data.GetHostList()) > 0 {
		hosts := make([]ui.Host, 0)
		containersMap := make(map[string][]ui.Container)

		for _, hostList := range data.GetHostList() {
			for _, hostInfo := range hostList.GetHosts() {
				// Convert host info
				host := ConvertHostInfo(hostInfo)
				hosts = append(hosts, host)

				// Convert container info for this host
				containers := make([]ui.Container, 0)
				for _, containerInfo := range hostInfo.GetContainer() {
					container := ConvertContainerInfo(containerInfo)
					containers = append(containers, container)
				}
				containersMap[host.MACAddress] = containers
			}
		}

		// Update UI on main thread
		dm.app.QueueUpdateDraw(func() {
			dm.app.UpdateHosts(hosts)
			dm.app.UpdateContainersMap(containersMap)
		})
	}

	// Update services status
	if len(data.GetServicesStatus()) > 0 {
		service := ConvertServicesStatus(data.GetServicesStatus())
		service.TotalHosts = len(data.GetHostList())

		dm.app.QueueUpdateDraw(func() {
			dm.app.UpdateServices(service)
		})
	}

	// Update logs
	if data.GetLogs() != "" {
		dm.app.QueueUpdateDraw(func() {
			dm.app.UpdateLogs(data.GetLogs())
		})
	}

	// Update cron time
	if data.GetCronTime() > 0 {
		dm.app.QueueUpdateDraw(func() {
			dm.app.UpdateCronTime(data.GetCronTime())
		})
	}
}

// sendHeartbeats sends periodic acknowledgments to keep stream alive
func (dm *DataManager) sendHeartbeats(stream orchestrator.TUIService_SendDatastreamClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			err := stream.Send(&orchestrator.DataStreamreq{
				Ack: "HEARTBEAT",
			})
			if err != nil {
				log.Printf("Error sending heartbeat: %v", err)
				return
			}
		}
	}
}

func (dm *DataManager) Stop() {
	dm.cancel()
}
