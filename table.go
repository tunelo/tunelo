/*
 * This file is part of YAVA (YAVA, Another VPN Application).
 *
 *
 * YAVA is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
 *
 * YAVA is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
 * of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with YAVA. If not, see <http://www.gnu.org/licenses/>.
 *
 * Author: Emiliano A. Billi emiliano.billi@gmail.com
 * Date: 2024
 */

package tunelo

import (
	"encoding/binary"
	"net"
)

type RouteTable struct {
	table map[uint32]uint16
	defgw uint16
}

func (r *RouteTable) SetDestination(ip net.IP, vaddr uint16) {
	if r.table == nil {
		r.table = make(map[uint32]uint16)
	}
	ipb := binary.BigEndian.Uint32(ip.To4())
	r.table[ipb] = vaddr
}
func (r *RouteTable) GetDestination(ip net.IP) (uint16, bool) {
	if r.table == nil {
		r.table = make(map[uint32]uint16)
	}
	vaddr, ok := r.table[binary.BigEndian.Uint32(ip.To4())]
	return vaddr, ok
}
