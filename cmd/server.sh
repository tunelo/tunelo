#!/bin/bash

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
source "$CONFIG_FILE"

# Ruta del ejecutable de Tunelo VPN Server
TUNELO_EXEC="./tunelo"

# Función para iniciar el servidor como daemon
function start_server() {
    if [ -f "$PID_FILE" ]; then
        echo "Tunelo VPN Server is running"
        exit 1
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
    nohup $TUNELO_EXEC -mode=server -sudp_config "$SUDP_CONFIG" -utun_vaddr "$UTUN_VADDR" \
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
