#!/bin/sh

# External services are needed by current project, some of which could be started via Docker images.
# This script helps user to start configured services via simple command args.

USAGE="$(basename "$0") CMD | SERVS [SERVS...] -- script for 'services/' usage, make sure to execute from project root directory"
USAGE_CMD="CMD: stop"
USAGE_SERVS="SERVS: es|elastic, nsq, pg|postgre"
USAGE_EX="Example: $(basename "$0") pg | Start postgre service"
NETWORK="search-net"

NETWORK_INIT=0
init_docker_network() {
    if [ $NETWORK_INIT -eq 1 ]; then
        return
    fi
    echo "Inspecting docker network..."
    sudo docker network inspect $NETWORK > /dev/null
    if [ "$?" -eq 1 ]; then
        echo "Network '$NETWORK' does not exists, creating..."
        sudo docker network create -d bridge --attachable $NETWORK
    fi
    NETWORK_INIT=1
}

stop_docker_network() {
    sudo docker network rm $NETWORK
}

run_docker() {
    init_docker_network
    echo "Running $1..."
    sudo docker-compose -f services/$2 -p $2 up -d
}

stop_docker() {
    echo "Stopping $1..."
    sudo docker-compose -f services/$2 -p $2 down
}

SUCCESS=0
for ARG in "$@"; do
    case "$ARG" in
    es|elastic)
        run_docker elastic elastic-7.yml
        SUCCESS=1
        ;;
    nsq)
        run_docker nsqlookupd nsqlookupd-latest.yml
        run_docker nsqd nsqd-latest.yml
        run_docker nsqadmin nsqadmin-latest.yml
        SUCCESS=1
        ;;
    pg|postgres)
        run_docker postgres postgres.yml
        SUCCESS=1
        ;;
    stop)
        echo "Stopping containers..."
        stop_docker elastic elastic-7.yml
        stop_docker nsqlookupd nsqlookupd-latest.yml
        stop_docker nsqd nsqd-latest.yml
        stop_docker nsqadmin nsqadmin-latest.yml
        stop_docker postgres postgres-alpine.yml
        echo "Stopping network"
        stop_docker_network
        SUCCESS=1
        ;;
    *)
        printf "invalid arg '%s'\n" $ARG
        SUCCESS=0
        break
        ;;
    esac
done

if [ $SUCCESS -eq 0 ]; then
    echo $USAGE
    echo $USAGE_CMD
    echo $USAGE_SERVS
    echo $USAGE_EX
    exit 1
fi

exit 0
