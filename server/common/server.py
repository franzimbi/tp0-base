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
        self._agents_waiting = {}

    def graceful_shutdown(self, signum, frame):
        if self._server_socket:
            self._server_socket.close()
            logging.info("action: close_socket | result: success")
        for i in self._agents_waiting.values():
            if i:
                i.close()
                logging.info("action: agent_waiting_closed | result: success")
        exit(0)

    def run(self, agentsCount):
        """
        Dummy Server loop
        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        # the server
        signal.signal(signal.SIGTERM, self.graceful_shutdown)
        while True:
            logging.debug("SIZE OF AGENTS WAITING: " + str(len(self._agents_waiting)))
            logging.debug("AGENTS COUNT: " + str(agentsCount))
            if len(self._agents_waiting) == agentsCount:
                logging.debug("action: all_agents_connected | result: success")
                self.lottery(agentsCount)
                self._agents_waiting = {}
                logging.info(f"action: sorteo | result: success")
            protocol = self.__accept_new_connection()
            self.__handle_client_connection(protocol)
            # protocol.close()
    
    def lottery(self, agentsCount):
        results = {}
        for i in range(agentsCount):
            results[i+1] = []
        for bet in utils.load_bets(): # arma un arreglo con todos los dnis ganadores de cada agencia
            if utils.has_won(bet):
                results[bet.agency].append(bet.document)
        for id, protocol in self._agents_waiting.items(): # envia a cada agencia su array de dnis ganadores
            try:
                protocol.send_lottery_results(results[id])
                logging.debug(f"action: send_lottery_results_to_agent_{id} | result: in_progress | winners: {len(results[id])}")
            except (socket.error, OSError) as e:
                logging.error(f"action: send_lottery_results_to_agent_{id} | result: fail | error: {e}")
                continue
            logging.info(f"action: send_lottery_results_to_agent_{id} | result: success | winners: {len(results[id])}")
        for protocol in self._agents_waiting.values(): # cierra todas las conexiones
            if not protocol.recv_ack():
                logging.error("action: receive_ack | result: fail | error: Client disconnected")
            protocol.close()


    def __handle_client_connection(self, protocol):

        code = protocol.recv_code()
        if code is None:
            logging.error("action: receive_code | result: fail | error: Client disconnected")
            protocol.close()
            return
        agent_id = protocol.recv_agent_id()
        if agent_id is None:
            logging.error("action: receive_agent_id | result: fail | error: Client disconnected")
            protocol.close()
            return
        if code == b'\x02':  # Finish code
            logging.info("action: agent_waiting_result | result: success")
            self._agents_waiting[agent_id] = protocol

        if code == b'\x01': # Chunk code
            self.recv_chunck(protocol, agent_id)
       
 

    def recv_chunck(self, protocol, agent_id):
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
