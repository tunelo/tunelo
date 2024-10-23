/*
 * This file is part of Tunelo (Tunelo, Another VPN Application).
 *
 *
 * Tunelo is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
 *
 * Tunelo is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
 * of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with Tunelo. If not, see <http://www.gnu.org/licenses/>.
 *
 * Author: Emiliano A. Billi emiliano.billi@gmail.com
 * Date: 2024
 */

package tunelo

import (
	"fmt"
	"net"

	"github.com/tunelo/sudp"
	"github.com/tunelo/utun"

	"golang.org/x/net/ipv4"
)

type VnetSwitch struct {
	self    net.IP
	Tun     *utun.Utun
	sock    *sudp.ServerConn
	route   *RouteTable
	network *net.IPNet
}

func NewVnetSwitch(cird string, peer string, serverConfig string) (*VnetSwitch, error) {
	laddr, raddrs, e := sudp.ParseConfig(serverConfig)
	if e != nil {
		return nil, e
	}

	server, e := sudp.Listen(laddr, raddrs)
	if e != nil {
		return nil, e
	}
	iface, e := opentun(cird, peer)
	if e != nil {
		return nil, e
	}

	self, n, e := net.ParseCIDR(cird)

	vnets := VnetSwitch{
		self:    self,
		Tun:     iface,
		sock:    server,
		route:   &RouteTable{},
		network: n,
	}
	return &vnets, nil
}

func (v *VnetSwitch) Run() error {
	var (
		n int
		e error
	)
	err := make(chan error)
	buf := make([]byte, v.Tun.MTU)

	go func() {
		for {
			buff, vaddr, e := v.sock.RecvFrom()
			if e != nil {
				fmt.Println("warn sudp - RecvFrom() ", e, " - drop -")
				continue
			}
			ip, e := ipv4.ParseHeader(buff)
			if e != nil {
				fmt.Println("warn parsing header - ", e, " - drop -")
				continue
			}

			if !v.network.Contains(ip.Src) {
				fmt.Println(fmt.Sprintf("warn route from %s not enabled - drop -", ip.Src.String()))
				continue
			}

			v.route.SetDestination(ip.Src, vaddr)
			if vaddr, ok := v.route.GetDestination(ip.Dst); !ok || ip.Dst.Equal(v.self) {
				if _, e := v.Tun.Write(buff); e != nil {
					err <- fmt.Errorf("tun write: %v", e)
					return
				}
			} else {
				if e := v.sock.SendTo(buff, vaddr); e != nil {
					fmt.Println(fmt.Sprintf("warn sending to %d, %v", vaddr, e))
					continue
				}
			}
		}
	}()
	for {
		n, e = v.Tun.Read(buf)
		if e != nil {
			e = fmt.Errorf("tun read: %v", e)
			break
		}
		ip, e := ipv4.ParseHeader(buf)
		if e != nil {
			fmt.Println("warn parsing header - ", e, " - drop -")
			continue
		}
		vaddr, ok := v.route.GetDestination(ip.Dst)
		if !ok {
			fmt.Println(fmt.Sprintf("warn route not found to %s - drop -", ip.Src.String()))
			continue
		}
		if e := v.sock.SendTo(buf[0:n], vaddr); e != nil {
			fmt.Println(fmt.Sprintf("warn sending to %d, %v", vaddr, e))
			continue
		}
	}
	v.sock.Close()
	ge := <-err
	return fmt.Errorf("%v, %v", ge, e)
}
