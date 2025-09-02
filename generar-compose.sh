#!/bin/bash

if [ $# -ne 2 ]; then
    echo "El script debe recibir por parametro el nombre del archivo de salida y la cantidad de clientes esperados"
    echo "Ej: ./generar-compose.sh docker-compose-dev.yaml 5"
    exit 1
fi

echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"

echo "name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - CLIENTS_NUMBER=$2
    volumes:
      - ./server/config.ini:/config.ini
    networks:
      - testing_net
    " > $1

for i in $(seq 1 $2); do 
    echo "  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - NOMBRE=Santiago
      - APELLIDO=Lorca
      - DOCUMENTO=30904465
      - NACIMIENTO=1999-03-17
      - NUMERO=7574
    volumes:
      - ./client/config.yaml:/config.yaml
      - ./.data/agency-$i.csv:/agency.csv
    networks:
      - testing_net
    depends_on:
      - server
    " >> $1; done

echo "networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24" >> $1