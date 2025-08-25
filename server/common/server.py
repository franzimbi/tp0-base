import socket
import logging
import signal
import common.utils as utils
from common.protocol import Protocol

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def graceful_shutdown(self, signum, frame):
        if self._server_socket:
            self._server_socket.close()
            logging.info("action: close_socket | result: success")
        exit(0)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        signal.signal(signal.SIGTERM, self.graceful_shutdown)
        while True:
            protocol = self.__accept_new_connection()
            self.__handle_client_connection(protocol)
    

    # def recv_all(self, socket, size):
    #     buf = b''
    #     while len(buf) < size:
    #         n = socket.recv(size - len(buf))
    #         if n == 0:
    #             return None
    #         buf += n
    #     return buf

    # def recv_int(self, socket):
    #     buf = self.recv_all(socket, 4)
    #     if buf is None:
    #         return None
    #     return int.from_bytes(buf, byteorder='little')
    
    # def recv_string(self, socket):
    #     size = self.recv_all(socket, 1)
    #     if size is None:
    #         return None
    #     size = int.from_bytes(size, byteorder='little')

    #     string = self.recv_all(socket, size)
    #     if string is None:
    #         return None
    #     return string.decode('utf-8')

    def __handle_client_connection(self, protocol):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        # logging.info('action: recv_apuesta | result: in_progress')
        try:
            nombre, apellido, documento, nacimiento, numero = protocol.recv_bet()

            bet = utils.Bet("1", nombre, apellido, str(documento), nacimiento, str(numero))
            utils.store_bets([bet])

            logging.info(f'action: apuesta_almacenada | result: success | dni: {documento} | numero: {numero}')

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            protocol.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        protocol = Protocol(c)
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return protocol
