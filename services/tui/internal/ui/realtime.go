package ui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
)

// startRealtime starts both data and log streaming goroutines.
func (a *App) startRealtime() {
	if a.client == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	go a.runDataStream(ctx)
	go a.runLogStream(ctx)
}

func (a *App) runDataStream(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		stream, err := a.client.SendDatastream(ctx)
		if err != nil {
			a.logs.AddLog("[red]datastream connect failed: " + err.Error())
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		a.logs.AddLog("[green]datastream connected")
		stopHB := make(chan struct{})
		go func() {
			var n int64
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					atomic.AddInt64(&n, 1)
					_ = stream.Send(&tui.DataStreamReceived{Ack: fmt.Sprintf("hb-%d", atomic.LoadInt64(&n))})
				case <-stopHB:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
		for {
			msg, err := stream.Recv()
			if err != nil {
				close(stopHB)
				a.logs.AddLog("[red]datastream recv error: " + err.Error())
				break
			}
			a.handleSnapshot(msg)
		}
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (a *App) handleSnapshot(msg *tui.DataStreamSend) {
	if msg == nil {
		return
	}
	var hosts []Host
	containersMap := make(map[string][]Container)
	nameToMAC := make(map[string]string)
	if msg.HostList != nil {
		for _, h := range msg.HostList.Hosts {
			if h == nil {
				continue
			}
			tm, _ := time.Parse(time.RFC3339, h.LastHeartbeat)
			hosts = append(hosts, Host{Name: h.Hostname, IP: h.IpAddress, MACAddress: h.MacAddress, LastHeartbeat: tm})
			nameToMAC[h.Hostname] = h.MacAddress
			for _, c := range h.Containers {
				if c == nil {
					continue
				}
				containersMap[h.MacAddress] = append(containersMap[h.MacAddress], Container{
					Name:       c.Name,
					Image:      c.Image,
					Status:     protoStatusToString(c.Status),
					IsWatching: c.Watch,
					IsUpdating: false,
				})
			}
		}
	}
	svc := Service{}
	for _, s := range msg.ServicesStatus {
		switch s.ServicesStatus {
		case tui.ServicesStatus_ORCHESTRATOR:
			svc.OrchestratorStatus = s.Status
		case tui.ServicesStatus_REGISTRY_Monitor:
			svc.RegistryMonitorStatus = s.Status
		case tui.ServicesStatus_Database:
			svc.DatabaseStatus = s.Status
		}
	}
	svc.TotalHosts = len(hosts)
	cron := msg.CronTime

	a.dataMu.Lock()
	a.hostsData = hosts
	a.containersMap = containersMap
	a.nameToMAC = nameToMAC
	a.dataMu.Unlock()

	a.QueueUpdateDraw(func() {
		if len(hosts) > 0 {
			a.hosts.Update(hosts)
		}
		a.servicesStatus.Update(svc)
		a.cron.UpdateTime(cron)
		selected := a.hosts.selectedHostName
		if selected != "" {
			mac := nameToMAC[selected]
			a.containers.Update(containersMap[mac])
		}
	})

	if msg.Logs != "" {
		for _, line := range strings.Split(msg.Logs, "\n") {
			if line != "" {
				a.logs.AddLog(line)
			}
		}
	}
}

func protoStatusToString(st tui.ContainerInfo_Status) string {
	switch st {
	case tui.ContainerInfo_RUNNING:
		return "Running"
	case tui.ContainerInfo_STOPPED:
		return "Stopped"
	case tui.ContainerInfo_PAUSED:
		return "Paused"
	case tui.ContainerInfo_RESTARTING:
		return "Restarting"
	case tui.ContainerInfo_EXITED:
		return "Exited"
	case tui.ContainerInfo_DEAD:
		return "Dead"
	default:
		return "Unknown"
	}
}

func (a *App) runLogStream(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		stream, err := a.client.StreamLogs(ctx)
		if err != nil {
			a.logs.AddLog("[red]logstream connect failed: " + err.Error())
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		a.logs.AddLog("[green]logstream connected")
		stop := make(chan struct{})
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_ = stream.Send(&tui.DataStreamReceived{Ack: "log-hb"})
				case <-stop:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
		for {
			line, err := stream.Recv()
			if err != nil {
				close(stop)
				a.logs.AddLog("[red]logstream recv error: " + err.Error())
				break
			}
			if line != nil && line.Line != "" {
				a.logs.AddLog(line.Line)
			}
		}
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}
