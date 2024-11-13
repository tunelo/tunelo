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

package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/tunelo/sudp"
	"github.com/tunelo/tunelo"
	"github.com/tunelo/utun"
)

var (
	version = "0.1.2-alpha"
)

func main() {
	var (
		pri    *ecdsa.PrivateKey
		pub    *ecdsa.PublicKey
		raddr  *net.UDPAddr
		cidr   string
		peer   string
		hmac   string
		vaddr  int
		peergw bool
		mode   string
		config string
		ver    bool
		e      error
	)

	flag.IntVar(&vaddr, "sudp_vaddr", -1, "SUDP Virtual Address (e.g., 1001)")
	flag.StringVar(&mode, "mode", "client", "VPN mode: client | server")
	flag.StringVar(&config, "sudp_config", "", "SUDP server config file")
	flag.StringVar(&hmac, "sudp_hmac_key", "", "SUDP header hmac")
	flag.Func("sudp_pri", "Path to the SUDP Self Private key file in PEM format (e.g., private.prm)", func(s string) error {
		pri, e = sudp.PrivateFromPemFile(s)
		if e != nil {
			return e
		}
		return nil
	})

	flag.Func("sudp_pub", "Path to the SUDP Server's public key in PEM format (e.g., public.pem)", func(s string) error {
		pub, e = sudp.PublicKeyFromPemFile(s)
		if e != nil {
			return e
		}
		return nil
	})

	flag.Func("sudp_endpoint", "SUDP Server's address (e.g., 18.221.232.10:7000)", func(s string) error {
		raddr, e = net.ResolveUDPAddr("udp4", s)
		if e != nil {
			return e
		}
		return nil
	})

	flag.Func("utun_vaddr", "Virtual utun (osx) - tun/tap (linux) interface address in CIDR format (e.g., 10.0.0.2/24)", func(s string) error {
		_, _, e := net.ParseCIDR(s)
		if e != nil {
			return e
		}
		cidr = s
		return nil
	})

	flag.Func("utun_peer", "Peer vitual address (e.g., 10.0.0.1)", func(s string) error {
		e := net.ParseIP(s)
		if e == nil {
			return fmt.Errorf("invalid peer_address %s", s)
		}
		peer = s
		return nil
	})

	flag.BoolVar(&peergw, "peer_gw", false, "Set true if peer is the new default gw")

	flag.BoolVar(&ver, "version", false, "Show Tunelo and SUDP version")

	prefix := flag.String("keygen", "", "Create a Private/Public key pair in PEM format (e.g., -keygen <prefix>) ")

	flag.Parse()

	if ver {
		fmt.Printf("Tunelo: v%s - SUDP: %s\n", version, sudp.Version())
		os.Exit(2)
	}

	if *prefix != "" {
		pri := fmt.Sprintf("%s_private.pem", *prefix)
		pub := fmt.Sprintf("%s_public.pem", *prefix)
		e := sudp.GeneratePEMKeyPair(pri, pub)
		if e != nil {
			fmt.Errorf("generating key pair: %v", e)
			os.Exit(2)
		}
		fmt.Println(fmt.Sprintf("Success: Private Key: %s, Public Key: %s", pri, pub))
		os.Exit(0)
	}

	switch mode {
	case "client":
		var sharedHmac []byte
		if pri == nil {
			fmt.Println("SUDP Self Private key is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if pub == nil {
			fmt.Println("SUDP Server's Public key is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if raddr == nil {
			fmt.Println("SUDP Server address is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if peer == "" {
			fmt.Println("Peer virtual network address is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if cidr == "" {
			fmt.Println("Self Utun CIDR is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if vaddr == -1 {
			fmt.Println("SUDP virtual address is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		if vaddr == 0 {
			fmt.Println("SUDP virtual address 0 is a wrong value, 0 is reserver to the server")
			flag.Usage()
			os.Exit(2)
		}

		if hmac == "" {
			sharedHmac = nil
		} else {
			sharedHmac = []byte(hmac)
		}

		s, _ := net.ResolveUDPAddr("udp4", "0.0.0.0:")
		laddr := sudp.LocalAddr{
			VirtualAddress: uint16(vaddr),
			NetworkAddress: s,
			PrivateKey:     pri,
		}
		paddr := sudp.RemoteAddr{
			VirtualAddress: 0,
			NetworkAddress: raddr,
			SharedHmacKey:  sharedHmac,
			PublicKey:      pub,
		}
		c, err := tunelo.NewVnetClient(cidr, peer, &laddr, &paddr)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(c.Run())
	case "server":
		if config == "" {
			fmt.Println("Missing config file")
			flag.Usage()
			os.Exit(2)
		}

		if cidr == "" {
			fmt.Println("Self Utun CIDR is not present in the argument list")
			flag.Usage()
			os.Exit(2)
		}

		v, err := tunelo.NewVnetSwitch(cidr, utun.NOPEER, config)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(v.Run())
	default:
		fmt.Println("Invalid mode")
		flag.Usage()
		os.Exit(2)
	}

}
