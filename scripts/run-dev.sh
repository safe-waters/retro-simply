#!/usr/bin/env bash

set -euo pipefail

if [ "${1}" = "build" ] || [ "${1}" = "up" ] || [ "${1}" = "down" ] || [ "${1}" = "down-volumes" ]; then
    export DOCKER_BUILDKIT=1

    if [[ "${1}" == "build" ]]; then
        echo "building dev services"
        docker-compose --file docker-compose.yml --file docker-compose.dev.yml build
    elif [[ "${1}" == "up" ]]; then
        echo "building and running dev services"
        docker-compose --file docker-compose.yml --file docker-compose.dev.yml up -d --build --remove-orphans
    elif [[ "${1}" == "down" ]]; then
        echo "tearing down dev services"
        docker-compose --file docker-compose.yml --file docker-compose.dev.yml down --remove-orphans
    elif [[ "${1}" == "down-volumes" ]]; then
        echo "tearing down dev services and deleting all volumes"
        docker-compose --file docker-compose.yml --file docker-compose.dev.yml down --remove-orphans --volumes
    fi
else
    echo "incorrect argument - please use 'build', 'up', 'down', or 'down-volumes'"
    exit 1
fi
