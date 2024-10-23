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
	"github.com/tunelo/sudp"
	"github.com/tunelo/utun"

	"golang.org/x/net/ipv4"
)

const (
	udplen   = 8
	overhead = ipv4.HeaderLen + 8 + sudp.HeaderLen + sudp.DataHeaderLen
)

func mtu(m int) int {
	return m - overhead
}

func opentun(cird string, peer string) (*utun.Utun, error) {
	iface, err := utun.OpenUtun()
	if err != nil {
		return nil, err
	}

	if err = iface.SetMTU(mtu(1500)); err != nil {
		return nil, err
	}
	if err = iface.SetIP(cird, peer); err != nil {
		return nil, err
	}
	return iface, nil
}
