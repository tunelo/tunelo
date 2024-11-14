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

function wait_running() {
    local logfile=$1
    local peer=$2
    while true; do
        # Verifica si el archivo contiene alguna línea
        if [[ -f "$logfile" && -s "$logfile" ]]; then
            line=$(head -n 1 "$logfile")
            status=$(echo "$line" | grep -o 'status=[^ ]*' | sed 's/status=//')

            # Verificar si hay un mensaje de error
            if [[ "$status" == "error" ]]; then
                # Extraer el mensaje de error si existe
                message=$(echo "$line" | grep -o 'message=.*' | sed 's/message=//')
                echo "Status: $status, Message: $message"
                exit 1
            else
                echo "Status: $status"
                change_route default $peer
                rm -f "$temp_sudp_pri" "$temp_sudp_pub"
                (cleanup_daemon $pid)&
            fi
            break
        fi
        # Espera 1 segundo antes de volver a comprobar
        sleep 1
    done
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

cleanup_daemon() {
    pid_to_wait=$1
    while ps -p "$pid_to_wait" > /dev/null 2>&1; do sleep 2; done
    set_route default "$DEFAULT_GW"
}

check_config_file() {
    if [[ -z "$SUDP_ENDPOINT" || -z "$SUDP_PRI" || -z "$SUDP_PUB" || -z "$SUDP_VADDR" || -z "$UTUN_PEER" || -z "$UTUN_VADDR" ]]; then
        echo "Missing mandatory values in config file"
        exit 1
    fi
}

set_route() {
    local to=$1
    local target=$2
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sudo route add "$to" "$target" > /dev/null 2>&1
    else
        sudo ip route add "$to" via "$target" > /dev/null 2>&1
    fi
}

change_route() {
    local to=$1
    local target=$2
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sudo route change "$to" "$target" > /dev/null 2>&1
    else
        sudo ip route change "$to" via "$target" > /dev/null 2>&1
    fi
}

set_route_to_peer() {
    local default=$1
    if [[ $(get_route_to_peer) == $default ]]; then
        echo "Route to peer is OK: ($default -> $(get_ip_from_endpoint))"
    else
        echo "Setting route to peer: ($default -> $(get_ip_from_endpoint))"
        if [[ $(get_route_to_peer) == "" ]]; then
            set_route $(get_ip_from_endpoint) $default
        else
            change_route $(get_ip_from_endpoint) $default
        fi
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

    check_config_file

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
    set_route_to_peer $DEFAULT_GW
    
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
    pid=$!
    echo $pid > "$PID_FILE"
    echo "Tunelo VPN Client started with PID $(cat $PID_FILE)...waiting connection"
    wait_running $LOG_FILE $UTUN_PEER
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
