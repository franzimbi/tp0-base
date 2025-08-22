import sys

archivo_salida = sys.argv[1]
cantidad_clientes = int(sys.argv[2])

with open(archivo_salida, 'w') as f:
    f.write("name: tp0\nservices:\n")
    f.write("  server:\n    container_name: server\n    image: server:latest" \
    "\n    entrypoint: python3 /main.py" \
    "\n    environment:\n      - PYTHONUNBUFFERED=1\n      - LOGGING_LEVEL=DEBUG" \
    "\n    networks:\n      - testing_net\n\n")

    for i in range(1, cantidad_clientes +1):
        f.write(f"  client{i}:")
        f.write(f"\n    container_name: client{i}")
        f.write("\n    image: client:latest\n    entrypoint: /client\n    environment:\n")
        f.write(f"      - CLI_ID={i}")
        f.write("\n      - CLI_LOG_LEVEL=DEBUG\n    networks:\n      - testing_net\n    depends_on:\n      - server\n\n")

    f.write("networks:\n  testing_net:\n    ipam:\n      driver: default\n      config:\n        - subnet: 172.25.125.0/24\n")

