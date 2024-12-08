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
	"flag"
	"fmt"
	"os"

	"github.com/tunelo/sudp"
	"github.com/tunelo/tunelo"
	"github.com/tunelo/utun"
)

var (
	version = "0.1.3-alpha"
)

func main() {
	var (
		display bool
		mode    string
		config  string
		server  string
		public  string
		private string

		port   int
		cird   string
		client string
		peergw bool
		new    bool
		add    bool
		ver    bool
	)

	// Defines the operating mode of Tunelo.
	// Possible values:
	// - "client": Runs the VPN in client mode (default).
	// - "server": Runs the VPN in server mode.
	// - "config": Configuration mode, used in combination with -new or -add to manage configuration files.
	flag.StringVar(&mode, "mode", "client", "VPN operating mode: client | server | config (use with -new or -add)")

	// Specifies the configuration file for Tunelo.
	// This file contains the parameters required to initialize the system in any mode.
	flag.StringVar(&config, "config", "", "Path to the Tunelo configuration file")

	// Creates a new server configuration file.
	// Use this with -mode=config and specify the output file name using -server <filename.json>.
	flag.BoolVar(&new, "new", false, "Create a new server configuration file. Use with -mode=config and -server <filename.json>")

	// Adds a new peer (client) to an existing server configuration file and generates a client configuration file.
	// Use this with -mode=config and specify the server and client file paths with -server and -client.
	flag.BoolVar(&add, "add", false, "Add a new peer to the server configuration file and generate a client configuration file. -mode=config, -server and -client")

	// Displays the current versions of Tunelo and the SUDP protocol.
	// Useful for compatibility checks and ensuring the latest version is in use.
	flag.BoolVar(&ver, "version", false, "Show the current versions of Tunelo and SUDP")

	// Enables iterative mode to display real-time progress or debugging information.
	flag.BoolVar(&display, "iterative", false, "Enable iterative mode to display real-time information")

	// Specifies the client configuration file to use.
	// This file contains the client's virtual address and other client-specific parameters.
	flag.StringVar(&client, "client", "", "Path to the client configuration file (use with -add)")

	// Specifies the server configuration file to use.
	// This file contains the server's configuration, including peers and general server settings.
	flag.StringVar(&server, "server", "", "Path to the server configuration file")

	// Public IP address where the SUDP server listens for incoming connections.
	// This address must be reachable from the external network.
	flag.StringVar(&public, "public", "", "Public IP address where the SUDP server listens for external connections")

	// Private IP address where the SUDP server listens for internal connections.
	// This address is typically used for communication within the local network.
	flag.StringVar(&private, "private", "0.0.0.0", "Private IP address where the SUDP server listens for internal connections.")

	// Port number where the SUDP server listens for connections.
	// Both public and private addresses will use this port for communication.
	flag.IntVar(&port, "port", 7000, "Port number where the SUDP server listens.")

	// CIDR block for the tunnel's virtual IP address.
	// This address is assigned to the tunnel for communication.
	flag.StringVar(&cird, "cird", "10.0.0.1/24", "CIDR block for the tunnel's virtual IP address.")

	flag.BoolVar(&peergw, "peergw", false, "Use peer as default gateway. Warning - a route to SUDP server MUST exist")
	flag.Parse()

	if ver {
		fmt.Printf("Tunelo: v%s - SUDP: %s\n", version, sudp.Version())
		os.Exit(2)
	}

	if config == "" && (mode == "client" || mode == "server" || mode == "dump") {
		fmt.Println("Missing config file")
		flag.Usage()
		os.Exit(2)
	}

	switch mode {
	case "dump":
		cfg, e := tunelo.LoadClientConfig(config)
		if e != nil {
			fmt.Printf("status=error message=%v\n", e)
			os.Exit(2)
		}
		fmt.Printf("UTUN_PEER=\"%s\"\n", cfg.UtunPeer)
		fmt.Printf("UTUN_ADDR=\"%s\"\n", cfg.UtunAddr)
		if cfg.Sudp != nil {
			fmt.Printf("SUDP_ENDPOINT=\"%s\"\n", *cfg.Sudp.Server.NetworkAddress)
			fmt.Printf("SUDP_VADDR=\"%d\"\n", cfg.Sudp.Host.VirtualAddress)
		} else {
			fmt.Printf("SUDP_ENDPOINT=\"\"\n")
			fmt.Printf("SUDP_VADDR=\"\"\n")
		}
		return

	case "client":
		cfg, e := tunelo.LoadClientConfig(config)
		if e != nil {
			fmt.Printf("status=error message=%v\n", e)
			os.Exit(2)
		}

		c, err := tunelo.NewVnetClient(cfg.UtunAddr, cfg.UtunPeer, cfg.Sudp, peergw)
		if err != nil {
			fmt.Printf("status=error message=%v\n", err)
			os.Exit(2)
		} else {
			if display {
				c.Display()
			} else {
				fmt.Println("status=connected")
			}
		}
		err = c.Run()
		fmt.Printf("status=error message=%v\n", err)
	case "server":
		cfg, e := tunelo.LoadServerConfig(config)
		if e != nil {
			fmt.Printf("status=error message=%v\n", e)
			os.Exit(2)
		}
		v, err := tunelo.NewVnetSwitch(cfg.UtunAddr, utun.NOPEER, cfg.Sudp)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(v.Run())
	case "config":
		if !new && !add {
			fmt.Println("--mode=config must be used with --add or --new")
			os.Exit(2)
		}

		if new {
			if server == "" {
				fmt.Println("-new must be used with -server <filename.json>")
				os.Exit(2)
			}
			if public == "" {
				fmt.Println("-new must be used with -public <PUBLIC IP>")
				os.Exit(2)
			}
			cfg, err := tunelo.NewServerConfig(private, public, port, "10.0.0.1/24")
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			err = cfg.DumpServerConfig(server)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			fmt.Println("New server file:", server)
			os.Exit(0)
		}
		if add {
			if server == "" {
				fmt.Println("-add must be used with -server <filename.json>")
				os.Exit(2)
			}

			if client == "" {
				fmt.Println("-add must be used with -client <filename.json>")
				os.Exit(2)
			}
			cfg, err := tunelo.LoadServerConfig(server)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}

			peer, err := cfg.AddPeer()
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}

			err = peer.DumpClientConfig(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}

			err = cfg.DumpServerConfig(server)
			if err != nil {
				fmt.Println(err)
				os.Exit(2)
			}
			fmt.Println("Updated server file:", server)
			fmt.Println("New client file:", client)
		}

	default:
		fmt.Println("Invalid mode")
		flag.Usage()
		os.Exit(2)
	}
}
