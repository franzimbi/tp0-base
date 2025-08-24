#!/bin/bash
puerto=12345
mensaje="hoy juega boca"
red="tp0_testing_net"

respuesta=$(docker run --rm --network "$red" alpine sh -c "echo '$mensaje' | nc server $puerto")

if [ "$respuesta" == "$mensaje" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi