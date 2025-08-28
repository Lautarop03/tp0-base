#!/bin/bash

NETWORK="testing_net"
SERVER_IP="SERVER"
SERVER_PORT=12345
MESSAGE="MESSAGE"

RESPONSE=$(docker run --rm --network $NETWORK busybox sh -c "echo $MESSAGE | nc -w 3 $SERVER_IP $SERVER_PORT")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi