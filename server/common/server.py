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
        # the server
        signal.signal(signal.SIGTERM, self.graceful_shutdown)
        while True:
            protocol = self.__accept_new_connection()
            self.__handle_client_connection(protocol)
    

    def __handle_client_connection(self, protocol):
        """
        Read message from a specific client protocol and closes the protocol

        If a problem arises in the communication with the client, the
        client protocol will also be closed
        """
        try:
            agency_id = str(protocol.recv_agency_id())
            while True:
                logging.info("arranca el while true de recibir apuestas")
                try:
                    logging.info("esperando code")
                    code = protocol.recv_code_command()
                    logging.info(f"code recibido: {code}")
                    if code is None:
                        logging.error("code is none, client disconnected")
                        return
                    if code == b'\x01':
                        logging.info("action: fin_de_envio_de_apuestas | result: success")
                        protocol.close()
                        return
                    bets = protocol.recv_bets(agency_id)
                    utils.store_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    protocol.send_ack()
                except OSError as e:
                    logging.error(f"action: socket_closed | result: {e}")
                    break
        finally:
            logging.info("cerrando socket del cliente")
            # protocol.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        protocol = Protocol(c)
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return protocol
