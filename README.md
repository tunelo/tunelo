
# Tunelo Product Overview

**Tunelo** is a powerful networking tool designed to create secure, private **virtual LANs** (Local Area Networks) over the Internet. By utilizing Tunelo, users can connect remote servers and clients seamlessly, enabling them to communicate as if they were part of the same local network, regardless of their physical locations.

## Key Features:
- **Virtual LAN Creation**: Tunelo allows the establishment of a virtual network where clients and servers can interact with unique virtual addresses and private IPs, facilitating secure and efficient communication.
- **Gateway Functionality**: The Tunelo server can act as an Internet gateway, providing connected clients with controlled access to external networks while maintaining the security of the virtual environment.
- **Flexible Configuration**: Both the server and client configurations are highly customizable, with options to specify virtual addresses, IPs, and encryption keys for enhanced security.
- **End-to-End Encryption**: Ensures data transmitted between the server and clients is protected, leveraging modern cryptographic standards for secure communication.
- **Cross-Platform Compatibility**: Tunelo supports various platforms, making it versatile for different network setups and user requirements.

## Use Cases:
- **Remote Team Collaboration**: Organizations can use Tunelo to create a shared virtual workspace for team members who are geographically dispersed, enabling seamless data sharing and collaboration.
- **Secure IoT Connectivity**: Tunelo can connect IoT devices securely across different locations, ensuring that communication between devices is protected and reliable.
- **Private Cloud Access**: Enables remote access to cloud resources as if they were part of the local network, with enhanced security measures and customizable routing.

## Benefits:
- **Enhanced Security**: By creating a private network overlay, Tunelo isolates traffic and ensures that communication remains encrypted and secure from external threats.
- **Improved Flexibility**: Users can configure their networks according to their needs, whether it's for simple peer-to-peer connections or complex multi-client setups.
- **Scalability**: Tunelo can easily scale as more clients need to be connected, supporting a range of use cases from small businesses to larger organizations.

**Tunelo** empowers users to build robust, secure, and efficient virtual networks that bridge the gap between physical locations, simplifying remote communication and collaboration.

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
