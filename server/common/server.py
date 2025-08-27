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
        agency_id = str(protocol.recv_agency_id())

        while True:
            try:
                bets, size_expected = protocol.recv_bets()
                if bets is None:
                    logging.info(f'action: client_disconnected | agency: {agency_id}')
                    break
                if size_expected != len(bets):
                    logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
                    protocol.send_errorApuesta()
                else:
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    protocol.send_ack()

                for bet in bets:
                    (nombre, apellido, documento, nacimiento, numero) = bet
                    bet = utils.Bet(agency_id, nombre, apellido, str(documento), nacimiento, str(numero))
                    utils.store_bets([bet])
                    logging.info(f'action: apuesta_almacenada | result: success | dni: {documento} | numero: {numero}')

            except OSError as e:
                logging.error(f"action: socket_closed | result: {e}")
                return
            finally:
                protocol.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        protocol = Protocol(c)
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return protocol
