#!/bin/bash

# Archivo para almacenar el PID
PID_FILE="/var/run/tunelo.pid"

# Archivo de log para redirigir la salida y los errores
LOG_FILE="/var/log/tunelo.log"

# Función para mostrar el uso del script
function show_usage() {
    echo "Uso: $0 {start|stop|restart} [config_file]"
    exit 1
}

# Verificar que se pasen al menos dos argumentos
if [ $# -lt 2 ]; then
    show_usage
fi

# Cargar el archivo de configuración
CONFIG_FILE=$2

# Ruta del ejecutable de Tunelo VPN Server
TUNELO_EXEC="./tunelo"

# Función para iniciar el servidor como daemon
function start_server() {
    if [ -f "$PID_FILE" ]; then
        pid=$(cat "$PID_FILE")
        # Comprobar si el proceso con el PID aún está activo
        if [[ $(ps -p "$pid" | grep "$pid") != "" ]]; then
            echo "Tunelo VPN Server is already running with PID $pid."
            exit 1
        else
            echo "PID file exists but the process is not running. Cleaning up."
            rm -f "$PID_FILE"
        fi
    fi

    echo "Starting Tunelo VPN Server as a daemon..."
    echo "-----------------------------" 
    echo " Manual Configuration Steps  " 
    echo "-----------------------------"
    echo "1. Enable MASQUERADE (eth0 is an example, use your physical interface)" 
    echo "  $ sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE"
    echo "2. Enable IP Forward" 
    echo "  $ sudo echo 1 > /proc/sys/net/ipv4/ip_forward"

    # Ejecutar Tunelo en segundo plano como daemon y redirigir salida y errores
    nohup $TUNELO_EXEC -mode=server -config "$CONFIG_FILE" \
        >> "$LOG_FILE" 2>&1 &

    echo $! > "$PID_FILE"
    echo "Tunelo VPN Server started with PID $(cat $PID_FILE)."
}

# Función para detener el servidor
function stop_server() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        echo "Stopping Tunelo VPN Server (PID: $PID)..."
        kill "$PID" && rm -f "$PID_FILE"
        echo "Tunelo VPN Server stopped."
    else
        echo "Tunelo VPN Server is not running."
    fi
}

# Función para reiniciar el servidor
function restart_server() {
    stop_server
    start_server
}

# Manejo de opciones start, stop y restart
case "$1" in
    start)
        start_server
        ;;
    stop)
        stop_server
        ;;
    restart)
        restart_server
        ;;
    *)
        show_usage
        ;;
esac
