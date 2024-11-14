#!/bin/bash

# Archivo para almacenar el PID
PID_FILE="/var/run/tunelo.pid"

# Archivo de log para redirigir la salida y los errores
LOG_FILE="/var/log/tunelo.log"


TUNELO_EXEC="./tunelo"

# Función para mostrar el uso del script
function show_usage() {
    echo "Usage: $0 {start [config_file]|stop}"
    exit 1
}

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

get_route_to_peer_macos() {
    netstat -rn | grep $(get_ip_from_endpoint) | awk '{print $2}' | head -n 1
}

get_route_to_peer_linux() {
    ip route | grep $(get_ip_from_endpoint) | awk '{print $3}' | head -n 1
}

get_route_to_peer() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        get_route_to_peer_macos;
    else
        get_route_to_peer_linux;
    fi
}


# Función para iniciar el servidor como daemon
function start_client() {

    if [ -f "$PID_FILE" ]; then
        pid=$(cat "$PID_FILE")
        # Comprobar si el proceso con el PID aún está activo
        if [[ $(ps -p "$pid" | grep "$pid") != "" ]]; then
            echo "Tunelo VPN Client is already running with PID $pid."
            exit 1
        else
            echo "PID file exists but the process is not running. Cleaning up."
            rm -f "$PID_FILE"
        fi
    fi
    echo "Starting Tunelo VPN..."
    echo "Config file: $1"
    CONFIG_FILE=$1
    source "$CONFIG_FILE"

    # Crea los archivos de clave publica y privada para pasar a Tunelo
    # Crear archivos temporales
    temp_sudp_pri=$(mktemp)
    temp_sudp_pub=$(mktemp)

    # Escribir las claves en los archivos temporales
    echo "$SUDP_PRI" > "$temp_sudp_pri"
    echo "$SUDP_PUB" > "$temp_sudp_pub"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        DEFAULT_GW=$(get_default_gw_macos)
    else
        DEFAULT_GW=$(get_default_gw_linux)
    fi
    echo "Current default gateway: $DEFAULT_GW"

    if [[ $(get_route_to_peer) == $DEFAULT_GW ]]; then
        echo "Route to peer is OK: ($DEFAULT_GW -> $(get_ip_from_endpoint))"
    else
        echo "Setting route to peer: ($DEFAULT_GW -> $(get_ip_from_endpoint))"
        if [[ $(get_route_to_peer) == "" ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                route -n add $(get_ip_from_endpoint) $DEFAULT_GW
            else
                ip route add $(get_ip_from_endpoint) via $DEFAULT_GW
            fi
        else
            if [[ "$OSTYPE" == "darwin"* ]]; then
                route -n change $(get_ip_from_endpoint) $DEFAULT_GW
            else
                ip route change $(get_ip_from_endpoint) via $DEFAULT_GW
            fi
        fi
    fi

    echo "-----------------------------"
    echo " Manual Configuration Steps  "
    echo "-----------------------------"
    echo "1. Change default gateway VPN endpoint"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "  $ sudo route change default" $UTUN_PEER
    else
        echo "  $ sudo ip route add default via " $UTUN_PEER
    fi

    echo "After disconnect, change default gateway at original configuration"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "  $ sudo route add default" $DEFAULT_GW
    else
        echo "  $ sudo ip route add default via " $DEFAULT_GW
    fi

    rm -f $LOG_FILE

    if [[ "$SUDP_HMAC_KEY" == "" ]]; then
        nohup $TUNELO_EXEC \
            -sudp_endpoint $SUDP_ENDPOINT \
            -sudp_pri $temp_sudp_pri \
            -sudp_pub $temp_sudp_pub \
            -sudp_vaddr $SUDP_VADDR \
            -utun_peer $UTUN_PEER \
            -utun_vaddr $UTUN_VADDR > "$LOG_FILE" 2>&1 &
    else
        nohup $TUNELO_EXEC \
            -sudp_endpoint $SUDP_ENDPOINT \
            -sudp_pri $temp_sudp_pri \
            -sudp_pub $temp_sudp_pub \
            -sudp_hmac_key $SUDP_HMAC_KEY \
            -sudp_vaddr $SUDP_VADDR \
            -utun_peer $UTUN_PEER \
            -utun_vaddr $UTUN_VADDR > "$LOG_FILE" 2>&1 &
    fi
    echo $! > "$PID_FILE"
    echo "Tunelo VPN Client started with PID $(cat $PID_FILE)."
}

function stop_client() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        echo "Stopping Tunelo VPN Client (PID: $PID)..."
        kill "$PID" && rm -f "$PID_FILE"
        echo "Tunelo VPN Client stopped."
    else
        echo "Tunelo VPN Client is not running."
    fi
}

# Manejo de opciones start, stop y restart
case "$1" in
    start)
        if [[ -z "$2" || ! -f "$2" ]]; then
            echo "Missing or invalid config file"
            show_usage
        fi
        start_client $2
        ;;
    stop)
        stop_client
        ;;
    *)
        show_usage
        ;;
esac
