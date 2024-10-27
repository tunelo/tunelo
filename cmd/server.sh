#!/bin/bash

# Cargar el archivo de configuración
source $1

echo "Starting Tunelo VPN Server..."

echo "-----------------------------"
echo " Manual Configuration Steps  "
echo "-----------------------------"
echo "1. Enable MASQUERADE (eth0 is and example, use your physical interface)"
echo "  $ iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE"
echo "2. Enable IP Forward"
echo "  $ echo 1 > /proc/sys/net/ipv4/ip_forward"


echo \
    -sudp_endpoint $SUDP_ENDPOINT \
    -sudp_pri $SUDP_PRI \
    -sudp_pub $SUDP_PUB \
    -sudp_vaddr $SUDP_VADDR \
    -utun_peer $UTUN_PEER \
    -utun_vaddr $UTUN_VADDR

