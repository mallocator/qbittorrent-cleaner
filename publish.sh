#!/usr/bin/env bash

docker build --platform linux/amd64 -t qbittorrent-cleaner .

docker image tag qbittorrent-cleaner mallox/qbittorrent-cleaner:latest
docker image tag qbittorrent-cleaner mallox/qbittorrent-cleaner:v1.0.1

docker push mallox/qbittorrent-cleaner:latest
docker push mallox/qbittorrent-cleaner:v1.0.1