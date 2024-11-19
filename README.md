
# Tunelo User Guide

This guide provides detailed instructions to set up and use **Tunelo** on both the **server** and the **client**.

---

## Server Configuration

### 1. Extract the package
Extract the Tunelo package to access the required files:
```bash
tar -vxzf tunelo.tar.gz
```

### 2. Navigate to the Tunelo directory
Move into the extracted directory:
```bash
cd tunelo
```

### 3. Configure the server
Generate the server configuration file (`server.json`) by running the following command:
```bash
./tunelo -mode config -new -server=server.json -public=<PUBLIC IP>
```
- **`<PUBLIC IP>`**: Replace with the public IP address of your server.

> **Note:** The server uses port **7000** by default. To specify a different port, use the `-port` option when creating the configuration file. Ensure that the firewall is configured to allow incoming UDP packets on the specified port.

### 4. Add a client
Add a client configuration to the server. This command creates a `client.json` file for the client and updates the `server.json` file with the client's information:
```bash
./tunelo -mode config -add -server=server.json -client=client.json
```

### 5. Start the server
Start the Tunelo server using the generated configuration file:
```bash
sudo ./server.sh start server.json
```

---

## Client Configuration

### 1. Extract the package
Extract the Tunelo package:
```bash
tar -vxzf tunelo.tar.gz
```

### 2. Navigate to the Tunelo directory
Move into the extracted directory:
```bash
cd tunelo
```

### 3. Add the client configuration file
Copy the `client.json` file created on the server to the client's machine and place it in the Tunelo directory.

### 4. Start the client
Start the Tunelo client using the configuration file:
```bash
sudo ./client.sh start client.json
```

### 5. Stop the client
To stop the Tunelo client, use the following command:
```bash
sudo ./client.sh stop
```

---

## Notes

- **Generated files**:
  - `server.json`: Contains the server configuration.
  - `client.json`: Contains the specific client configuration.
- Ensure that necessary ports are open on the server's firewall. By default, Tunelo uses **port 7000** for UDP traffic.
- Use superuser permissions (`sudo`) to start or stop Tunelo services.
- For troubleshooting, check the logs in the Tunelo directory.

---

With this guide, you can easily configure and use **Tunelo** on your server and client systems.
