import sys

archivo_salida = sys.argv[1]
cantidad_clientes = int(sys.argv[2])

def crear_server():
    return """  server:
    container_name: server
    image: server:latest
    volumes:
      - ./server/config.ini:/config.ini 
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
    networks:
      - testing_net
"""

def crear_cliente(id):
    return f"""  client{id}:
    container_name: client{id}
    image: client:latest
    volumes:
      - ./client/config.yaml:/config.yaml
    entrypoint: /client
    environment:
      - CLI_ID={id}
      - CLI_NOMBRE={'nombre' + str(id)}
      - CLI_APELLIDO={'apellido' + str(id)}
      - CLI_DOCUMENTO={1000 + id}
      - CLI_NACIMIENTO = {'1999-03-' + str(id)}
      - CLI_NUMERO={7110+id}
    networks:
      - testing_net
    depends_on:
      - server
"""

def crear_red():
    return """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

with open(archivo_salida, 'w') as f:
    f.write("name: tp0\nservices:\n")
    f.write(crear_server())
    for i in range(1, cantidad_clientes +1):
        f.write(crear_cliente(i) + "\n")
    f.write(crear_red())