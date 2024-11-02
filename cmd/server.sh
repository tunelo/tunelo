#!/bin/bash

# Cargar el archivo de configuración
source $1

echo "Starting Tunelo VPN Server..."

echo "-----------------------------"
echo " Manual Configuration Steps  "
echo "-----------------------------"
echo "1. Enable MASQUERADE (eth0 is and example, use your physical interface)"
echo "  $ sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE"
echo "2. Enable IP Forward"
echo "  $ sudo echo 1 > /proc/sys/net/ipv4/ip_forward"

./tunelo \
    -mode=server \
    -sudp_config $SUDP_CONFIG \
    -utun_vaddr  $UTUN_VADDR

