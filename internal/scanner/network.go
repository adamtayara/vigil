package scanner

import (
	"fmt"
	"net"
	"time"

	gopsnet "github.com/shirou/gopsutil/v4/net"
	gops "github.com/shirou/gopsutil/v4/process"
)

func ScanNetwork() ([]Connection, error) {
	conns, err := gopsnet.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("listing connections: %w", err)
	}

	pidNames := buildPIDNameMap()

	var out []Connection
	for _, c := range conns {
		conn := Connection{
			PID:        c.Pid,
			LocalAddr:  c.Laddr.IP,
			LocalPort:  c.Laddr.Port,
			RemoteAddr: c.Raddr.IP,
			RemotePort: c.Raddr.Port,
			Status:     c.Status,
			Family:     c.Family,
			Type:       c.Type,
		}
		if name, ok := pidNames[c.Pid]; ok {
			conn.ProcessName = name
		}
		out = append(out, conn)
	}
	return out, nil
}

func buildPIDNameMap() map[int32]string {
	m := make(map[int32]string)
	procs, err := gops.Processes()
	if err != nil {
		return m
	}
	for _, p := range procs {
		if name, err := p.Name(); err == nil {
			m[p.Pid] = name
		}
	}
	return m
}

func IsExternalIP(ip string) bool {
	if ip == "" || ip == "0.0.0.0" || ip == "::" {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return !parsed.IsLoopback() && !parsed.IsPrivate() && !parsed.IsLinkLocalUnicast()
}

func ReverseDNS(ip string) string {
	// 2-second timeout to avoid slowing down the scan
	type result struct {
		name string
	}
	ch := make(chan result, 1)
	go func() {
		names, err := net.LookupAddr(ip)
		if err != nil || len(names) == 0 {
			ch <- result{}
			return
		}
		ch <- result{names[0]}
	}()

	// Use a short timeout
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case r := <-ch:
		return r.name
	case <-timer.C:
		return ""
	}
}
