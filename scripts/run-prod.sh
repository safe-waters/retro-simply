#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

if [ "${1}" = "build" ] || [ "${1}" = "up" ] || [ "${1}" = "down" ] || [ "${1}" = "down-volumes" ]; then
    export DOCKER_BUILDKIT=1

    if [[ "${1}" == "down" ]]; then
        echo "tearing down production services"
        docker-compose --file docker-compose.yml --file docker-compose.prod.yml down --remove-orphans
        exit 0
    elif [[ "${1}" == "down-volumes" ]]; then
        echo "tearing down production services and volumes"
        docker-compose --file docker-compose.yml --file docker-compose.prod.yml down --remove-orphans --volumes
        exit 0
    fi

    (
        . .env
        cd frontend

        echo "packaging the frontend for production"
        docker build . --build-arg VUE_APP_API_VERSION=${API_VERSION} --output dist
        rm -rf ../server/dist
        mv dist ../server
    )

    if [[ "${1}" == "build" ]]; then
        echo "building production services"
        docker-compose --file docker-compose.yml --file docker-compose.prod.yml build
    elif [[ "${1}" == "up" ]]; then
        echo "building and running production services"
        docker-compose --file docker-compose.yml --file docker-compose.prod.yml up -d --build --remove-orphans
    fi
else
    echo "incorrect argument - please use 'build', 'up', 'down', or 'down-volumes'"
    exit 1
fi