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
            # protocol.close()
    

    def __handle_client_connection(self, protocol):
       
        agent_id = protocol.recv_agent_id()
        if agent_id is None:
            logging.error("action: receive_agent_id | result: fail | error: Client disconnected")
            protocol.close()
            return
        bets_size = protocol.recv_int()
        if bets_size is None:
            logging.error("action: receive_bets_size | result: fail | error: Client disconnected")
            protocol.close()
            return
        

        bets_procesados = 0
        for i in range(bets_size):
            try:
                nombre, apellido, documento, nacimiento, numero = protocol.recv_bets()

                bet = utils.Bet(str(agent_id), nombre, apellido, str(documento), nacimiento, str(numero))
                utils.store_bets([bet])
                bets_procesados += 1
                logging.debug(f"Processed bet {i+1}/{bets_size}")
            except OSError as e:
                logging.error("ction: apuesta_recibida | result: fail | cantidad: {bets_size}")

        logging.info(f'action: apuesta_recibida | result: success | cantidad: {bets_procesados}')
        protocol.send_ack()
        protocol.close()

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        protocol = Protocol(c)
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return protocol
