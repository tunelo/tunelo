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
)

type VnetClient struct {
	self net.IP
	Tun  *utun.Utun
	sock *sudp.ClientConn
}

func NewVnetClient(cird string, peer string, laddr *sudp.LocalAddr, raddr *sudp.RemoteAddr) (*VnetClient, error) {
	client, e := sudp.Connect(laddr, raddr)
	if e != nil {
		return nil, e
	}
	iface, e := opentun(cird, peer)
	if e != nil {
		return nil, e
	}

	self, _, e := net.ParseCIDR(cird)
	vnetc := VnetClient{
		self: self,
		Tun:  iface,
		sock: client,
	}
	return &vnetc, nil
}

func (c *VnetClient) Run() error {
	var (
		n int
		e error
	)
	err := make(chan error)
	buf := make([]byte, c.Tun.MTU)

	go func() {
		for {
			buff, e := c.sock.Recv()
			if e != nil {
				err <- fmt.Errorf("at sudp recv: %v", e)
				return
			}
			if _, e := c.Tun.Write(buff); e != nil {
				err <- fmt.Errorf("at tun write: %v", e)
				return
			}
		}
	}()

	for {
		n, e = c.Tun.Read(buf)
		if e != nil {
			break
		}
		if e = c.sock.Send(buf[:n]); e != nil {
			break
		}
	}
	return fmt.Errorf("%v, %v", c.sock.Close(), e)
}
