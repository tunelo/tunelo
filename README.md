
# Installation Guide for Tunelo

## Installing from Golang
To install Tunelo using Go, run:
```bash
$ go install github.com/tunelo/tunelo/cmd@latest
```

## Installing from TGZ
To install Tunelo from a compressed tarball:
```bash
$ tar -vxzf tunelo-v0.1-alpha_linux
```

## Generating Key Pairs
You need to generate key pairs for both the server and client.

Example:
```bash
$ ./tunelo -keygen server
```
This command creates a private/public key pair with the prefix server: server_public.pem and server_private.pem. Do the same for the client

## Server Instructions
Create a copy of the `sudp_config.json` file and modify it as needed:

### Example `sudp_config.json`
```json
{
  "server": {
    "virtual_address": 0,
    "listen": "0.0.0.0",
    "port": 7000,
    "private_key": "server_private.pem"
  },
  "peers": [
    {
      "virtual_address": 1001,
      "public_key": "client_public.pem"
    }
  ]
}
```
**Explanation**: The `virtual_address` of the peer is an arbitrary number that both the client and server must know. It functions similarly to a "port".

### Edit the Server Configuration File
Ensure the `server_example.conf` file is configured as follows:

```bash
# server_example.conf

# Virtual utun/tun interface address in CIDR format
UTUN_VADDR="10.0.0.1/24"

# Server SUDP Configuration file
SUDP_CONFIG="sudp_config.json"
```
**Note**: Adjust the CIDR format and configuration path as necessary for your environment.

## Client Configuration
Create and modify the `client_example.conf` file:

```bash
# client_example.conf

# Set to true if peer is the new default gateway
PEER_GW=true

# SUDP server's public address
SUDP_ENDPOINT="3.77.128.74:7000"

# Path to the SUDP Self Private key file in PEM format
SUDP_PRI="client_private.pem"

# Path to the SUDP Server's public key in PEM format
SUDP_PUB="server_public.pem"

# SUDP Virtual Address (integer)
SUDP_VADDR=1001

# Peer virtual address
UTUN_PEER="10.0.0.1"

# Virtual utun/tun/tap interface address in CIDR format
UTUN_VADDR="10.0.0.2/24"
```

## Starting the Server
To start the server, run:
```bash
$ sudo ./server.sh server_example.conf
```

## Starting the Client
To start the client, run:
```bash
$ sudo ./client.sh client_example.conf
```

Ensure that both the server and client configurations align with the network and security requirements of your deployment.
