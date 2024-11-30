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
function get_route_linux() {
    local target=${1:-default}  # Si no se proporciona $1, usa "default"
    ip route show | awk -v target="$target" '$0 ~ target {print $3}'
}

# Función para obtener el default gateway actual en macOS
function get_route_macos() {
    local target=${1:-default}
    via=$(route -n get $target | awk '/destination/ {print $2; exit}')
    if [[ ($via == "default" && $target == "default") || $via == $target ]]; then
        route -n get $target | awk '/gateway/ {print $2; exit}'
    fi

}

function get_route() {
    local target=$1
    if [[ "$OSTYPE" == "darwin"* ]]; then
        get_route_macos $target;
    else
        get_route_linux $target;
    fi
}

function set_route() {
    local to=$1
    local gw=$2
    if [[ "$OSTYPE" == "darwin"* ]]; then
        route add "$to" "$gw" > /dev/null 2>&1
    else
        ip route add "$to" via "$gw" > /dev/null 2>&1
    fi
}

function change_route() {
    local to=$1
    local gw=$2
    if [[ "$OSTYPE" == "darwin"* ]]; then
        route change "$to" "$gw" > /dev/null 2>&1
    else
        ip route change "$to" via "$gw" > /dev/null 2>&1
    fi
}

function get_ip_from_address() {
    local cird=$1
    echo "$cird" | cut -d':' -f1
}

function restart_default_gw() {
    local pid=$1
    local dgw=$2
    while ps -p "$pid" > /dev/null 2>&1; do sleep 2; done
    if [[ $(get_route) == "" ]]; then
        set_route default "$dgw"
    fi
}

function wait_running() {
    local file=$1
    while true; do
        if [[ -f "$file" && -s "$file" ]]; then
            status=$(head -n 1 "$file" | grep -o 'status=[^ ]*' | sed 's/status=//')
            if [[ "$status" == "error" ]]; then
                message=$(head -n 1 "$file" | grep -o 'message=.*' | sed 's/message=//')
                echo "Status: $status, Message: $message"
                exit 1
            else
                echo "Status: $status"
                exit 0
            fi
            break
        fi
        sleep 1
    done
}


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

    eval "$($TUNELO_EXEC --mode=dump --config=$CONFIG_FILE | grep -E '^[A-Za-z_][A-Za-z0-9_]*=.*$')"
    if [[ -z $SUDP_ENDPOINT ]]; then
        echo "missing SUDP_ENDPOINT"
        exit 1
    fi

    DEFAULTGW=$(get_route)
    SERVER_IP=$(get_ip_from_address $SUDP_ENDPOINT)
    ROUTEPEER=$(get_route $SERVER_IP)

    printf "%18s %s\n" "default gateway:" "$DEFAULTGW"
    printf "%18s %s\n" "tunelo server:" "$SERVER_IP"
    printf "%18s %s" "static to server:" "$ROUTEPEER"
    if [[ $ROUTEPEER == "" ]]; then
        printf "%s %s\n" "set to ->" "$DEFAULTGW"
        set_route $SERVER_IP $DEFAULTGW
    else
        if [[ $ROUTEPEER != $DEFAULTGW ]]; then
            printf "%s %s\n" "change to ->" "$DEFAULTGW"
            change_route $SERVER_IP $DEFAULTGW
        else
            printf "\n"
        fi
    fi

    rm -f $LOG_FILE
    nohup $TUNELO_EXEC -config $CONFIG_FILE > "$LOG_FILE" 2>&1 &
    pid=$!
    echo $pid > "$PID_FILE"
    echo "Tunelo VPN Client started with PID $pid...waiting connection"
    wait_running $LOG_FILE
    restart_default_gw $pid $DEFAULTGW &
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
