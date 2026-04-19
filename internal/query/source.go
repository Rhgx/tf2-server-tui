package query

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"tf2-server-tui/internal/config"
)

var a2sInfoRequest = append(
	[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x54},
	append([]byte("Source Engine Query"), 0x00)...,
)

type ServerInfo struct {
	IP         string
	Port       int
	Label      string
	Name       string
	Map        string
	Players    int
	MaxPlayers int
	Ping       int
	Online     bool
	Error      string
}

// QueryAll queries each configured server concurrently and returns one result
// per input server in the same order the config provided them, which keeps
// selection indices stable in the TUI.
func QueryAll(servers []config.ServerConfig) []ServerInfo {
	results := make([]ServerInfo, len(servers))
	var wg sync.WaitGroup

	for i, server := range servers {
		wg.Add(1)
		go func(index int, cfg config.ServerConfig) {
			defer wg.Done()
			results[index] = QueryServer(cfg)
		}(i, server)
	}

	wg.Wait()
	return results
}

// QueryServer sends a Source A2S_INFO query to a single TF2 server and
// normalizes the result into the ServerInfo shape used by the TUI. On failure,
// the returned ServerInfo still includes the original address information plus
// an Error message.
func QueryServer(server config.ServerConfig) ServerInfo {
	base := ServerInfo{
		IP:    server.IP,
		Port:  server.Port,
		Label: server.Label,
		Name:  displayName(server),
	}

	address := net.JoinHostPort(server.IP, fmt.Sprintf("%d", server.Port))
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		base.Error = err.Error()
		return base
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		base.Error = err.Error()
		return base
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		base.Error = err.Error()
		return base
	}

	start := time.Now()
	payload, err := queryInfoPayload(conn)
	if err != nil {
		base.Error = err.Error()
		return base
	}

	info, err := parseInfoPayload(payload)
	if err != nil {
		base.Error = err.Error()
		return base
	}

	base.Name = choose(info.Name, base.Name)
	base.Map = info.Map
	base.Players = info.Players
	base.MaxPlayers = info.MaxPlayers
	base.Ping = int(time.Since(start).Milliseconds())
	base.Online = true
	return base
}

type infoReply struct {
	Name       string
	Map        string
	Players    int
	MaxPlayers int
}

// queryInfoPayload handles the low-level A2S_INFO request/response flow. Some
// Source servers require a challenge round-trip before returning the actual
// info payload, and this helper hides that detail from QueryServer.
func queryInfoPayload(conn *net.UDPConn) ([]byte, error) {
	if _, err := conn.Write(a2sInfoRequest); err != nil {
		return nil, err
	}

	packet, err := readPacket(conn)
	if err != nil {
		return nil, err
	}

	if len(packet) < 5 {
		return nil, fmt.Errorf("short server response")
	}

	header := packet[4]
	switch header {
	case 0x49:
		return packet[5:], nil
	case 0x41:
		if len(packet) < 9 {
			return nil, fmt.Errorf("short challenge response")
		}
		request := make([]byte, 0, len(a2sInfoRequest)+4)
		request = append(request, a2sInfoRequest...)
		request = append(request, packet[5:9]...)
		if _, err := conn.Write(request); err != nil {
			return nil, err
		}

		packet, err = readPacket(conn)
		if err != nil {
			return nil, err
		}
		if len(packet) < 6 || packet[4] != 0x49 {
			return nil, fmt.Errorf("unexpected post-challenge response")
		}
		return packet[5:], nil
	default:
		return nil, fmt.Errorf("unexpected response header 0x%X", header)
	}
}

// readPacket reads and validates one Source query packet. Multi-packet replies
// are rejected because this lightweight browser keeps the query implementation
// intentionally small.
func readPacket(conn *net.UDPConn) ([]byte, error) {
	buffer := make([]byte, 1400)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	packet := append([]byte(nil), buffer[:n]...)

	if len(packet) >= 4 && bytes.Equal(packet[:4], []byte{0xFE, 0xFF, 0xFF, 0xFF}) {
		return nil, fmt.Errorf("multi-packet responses are not supported")
	}

	if len(packet) < 5 || !bytes.Equal(packet[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) {
		return nil, fmt.Errorf("invalid packet prefix")
	}

	return packet, nil
}

// parseInfoPayload extracts the A2S_INFO fields the UI actually displays. The
// parser intentionally stays small by skipping fields the UI does not use.
func parseInfoPayload(payload []byte) (infoReply, error) {
	reader := bytes.NewReader(payload)
	if _, err := reader.ReadByte(); err != nil {
		return infoReply{}, fmt.Errorf("missing protocol byte: %w", err)
	}

	name, err := readCString(reader)
	if err != nil {
		return infoReply{}, fmt.Errorf("missing server name: %w", err)
	}

	currentMap, err := readCString(reader)
	if err != nil {
		return infoReply{}, fmt.Errorf("missing map: %w", err)
	}

	if _, err := readCString(reader); err != nil {
		return infoReply{}, fmt.Errorf("missing folder: %w", err)
	}

	if _, err := readCString(reader); err != nil {
		return infoReply{}, fmt.Errorf("missing game name: %w", err)
	}

	var appID uint16
	if err := binary.Read(reader, binary.LittleEndian, &appID); err != nil {
		return infoReply{}, fmt.Errorf("missing app id: %w", err)
	}

	players, err := reader.ReadByte()
	if err != nil {
		return infoReply{}, fmt.Errorf("missing players: %w", err)
	}

	maxPlayers, err := reader.ReadByte()
	if err != nil {
		return infoReply{}, fmt.Errorf("missing max players: %w", err)
	}

	return infoReply{
		Name:       name,
		Map:        currentMap,
		Players:    int(players),
		MaxPlayers: int(maxPlayers),
	}, nil
}

func readCString(reader *bytes.Reader) (string, error) {
	buf := make([]byte, 0, 32)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			return string(buf), nil
		}
		buf = append(buf, b)
	}
}

func displayName(server config.ServerConfig) string {
	if server.Label != "" {
		return server.Label
	}
	return fmt.Sprintf("%s:%d", server.IP, server.Port)
}

func choose(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
