#!/bin/bash

# Cargar el archivo de configuración
source $1

# Función para obtener el default gateway actual en Linux
get_default_gw_linux() {
    ip route | grep default | awk '{print $3}' | head -n 1
}

# Función para obtener el default gateway actual en macOS
get_default_gw_macos() {
    netstat -rn | grep 'default' | awk '{print $2}' | head -n 1
}

# Función para obtener solo la dirección IP del SUDP_ENDPOINT (sin el puerto)
get_ip_from_endpoint() {
    echo "$SUDP_ENDPOINT" | cut -d':' -f1
}

get_ip_from_cidr() {
    echo "$UTUN_VADDR" | cut -d'/' -f1
}

if [[ "$OSTYPE" == "darwin"* ]]; then
    DEFAULT_GW=$(get_default_gw_macos)
else
    DEFAULT_GW=$(get_default_gw_linux)
fi
echo "Starting Tunelo VPN..."
echo "Current default gateway: $DEFAULT_GW"

echo "-----------------------------"
echo " Manual Configuration Steps  "
echo "-----------------------------"
echo "1. Set route to the endpoint: "

if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  $ sudo route -n add" $(get_ip_from_endpoint) $DEFAULT_GW
else
    echo "  $ sudo ip route add " $(get_ip_from_endpoint) " via " $DEFAULT_GW
fi
echo "2. Change default gateway VPN endpoint"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  $ sudo route change default" $UTUN_PEER
else
    echo "  $ sudo ip route add default via " $UTUN_PEER
fi

echo "3. To disconnet, change default gateway at original configuration"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  $ sudo route add default" $DEFAULT_GW
else
    echo "  $ sudo ip route add default via " $DEFAULT_GW
fi

if [[ "$SUDP_HMAC_KEY" == "" ]]; then
./tunelo \
    -sudp_endpoint $SUDP_ENDPOINT \
    -sudp_pri $SUDP_PRI \
    -sudp_pub $SUDP_PUB \
    -sudp_vaddr $SUDP_VADDR \
    -utun_peer $UTUN_PEER \
    -utun_vaddr $UTUN_VADDR
else
./tunelo \
    -sudp_endpoint $SUDP_ENDPOINT \
    -sudp_pri $SUDP_PRI \
    -sudp_pub $SUDP_PUB \
    -sudp_hmac_key $SUDP_HMAC_KEY \
    -sudp_vaddr $SUDP_VADDR \
    -utun_peer $UTUN_PEER \
    -utun_vaddr $UTUN_VADDR
fi


