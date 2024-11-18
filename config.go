package tunelo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"

	"github.com/tunelo/sudp"
)

type ServerConfig struct {
	UtunAddr string             `json:"utun_vaddr"`
	Mappings map[string]int     `json:"mappings"`
	Sudp     *sudp.ServerConfig `json:"sudp"`
}

type ClientConfig struct {
	UtunPeer string             `json:"utun_peer"`
	UtunAddr string             `json:"utun_vaddr"`
	Sudp     *sudp.ClientConfig `json:"sudp"`
}

func NewServerConfig(private string, public string, port int, utun string) (*ServerConfig, error) {
	server, err := sudp.NewServerConfig(private, public, port)
	if err != nil {
		return nil, err
	}

	_, _, err = net.ParseCIDR(utun)
	if err != nil {
		return nil, err
	}

	pri, pub := net.ParseIP(private), net.ParseIP(public)
	if pri == nil || pub == nil {
		return nil, fmt.Errorf("invalid public or private ip")
	}

	return &ServerConfig{
		UtunAddr: utun,
		Mappings: make(map[string]int),
		Sudp:     server,
	}, nil
}

func (s *ServerConfig) AddPeer() (*ClientConfig, error) {
	ip, _, err := net.ParseCIDR(s.UtunAddr)
	if err != nil {
		return nil, err
	}
	addr, _ := netip.ParseAddr(ip.String())
	maxv := 0

	for _, peer := range s.Sudp.Peers {
		if peer.VirtualAddress > maxv {
			maxv = peer.VirtualAddress
		}
	}
	maxv += 1
	for {
		next := addr.Next()
		addr = next
		if _, ok := s.Mappings[next.String()]; !ok {
			break
		}
	}

	peer, e := s.Sudp.AddPeer(maxv)
	if e != nil {
		return nil, e
	}
	s.Mappings[addr.String()] = maxv
	return &ClientConfig{
		UtunPeer: ip.String(),
		UtunAddr: fmt.Sprintf("%s/24", addr.String()),
		Sudp:     peer,
	}, nil
}

func LoadClientConfig(filePath string) (*ClientConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &ClientConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func LoadServerConfig(filePath string) (*ServerConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &ServerConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func (c *ClientConfig) DumpClientConfig(filePath string) error {
	content, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}

func (c *ServerConfig) DumpServerConfig(filePath string) error {
	content, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(content)
	return err
}
