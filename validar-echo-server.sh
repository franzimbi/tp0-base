#!/bin/bash
puerto=12345
mensaje="hoy juega boca"

docker exec -it -rm --network testing_net alpine sh

respuesta=$(echo "$mensaje" | nc server $puerto)
if ["$respuesta" == "$mensaje"]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi