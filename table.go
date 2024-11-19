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
	"encoding/binary"
	"net"
	"sync"
)

type RouteTable struct {
	lock  sync.RWMutex
	table map[uint32]uint16
	defgw uint16
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		table: make(map[uint32]uint16),
	}
}

func (r *RouteTable) SetDestination(ip net.IP, vaddr uint16) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ipb := binary.BigEndian.Uint32(ip.To4())
	r.table[ipb] = vaddr
}
func (r *RouteTable) GetDestination(ip net.IP) (uint16, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	vaddr, ok := r.table[binary.BigEndian.Uint32(ip.To4())]
	return vaddr, ok
}
