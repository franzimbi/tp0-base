import socket
import logging
import signal
import common.utils as utils
from common.utilsMonitor import MonitorUtils
from common.protocol import Protocol
from threading import Thread, Barrier, BrokenBarrierError

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._threads = []
        self._sockets = []
        self._barrier = None

    def graceful_shutdown(self, signum, frame):
        if self._server_socket:
            self._server_socket.close()
            logging.info("action: close_socket | result: success")
        for s in self._sockets:
            if s:
                s.close()
        logging.info("action: close_all_client_sockets | result: success")
        
        self._barrier.abort()
        for t in self._threads:
            t.join()
        logging.info("action: close_all_threads | result: success")
        exit(0)

    def run(self, agentsCount):
        """
        Dummy Server loop
        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        self._barrier = Barrier(agentsCount)
        # the server
        signal.signal(signal.SIGTERM, self.graceful_shutdown)
        while True:
            monitor = MonitorUtils()
            
            protocol = self.__accept_new_connection()
            thread = Thread(target=self.__handle_client_connection, args=(protocol, monitor, self._barrier))
            thread.start()
            self._threads = [t for t in self._threads if t.is_alive()]
            self._threads.append(thread)
            self._sockets = [s for s in self._sockets if s.is_alive()]
            self._sockets.append(protocol)

    def __handle_client_connection(self, protocol, monitor, barrier):

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
            logging.info("action: agent_waiting_result | result: success | agent_id: %d", agent_id)
            try:
                barrier.wait()
            except BrokenBarrierError:
                return
            winners_dnis = monitor.load_winners_thread_safe(agent_id)
            protocol.send_lottery_results(winners_dnis)
            if not protocol.recv_ack():
                logging.error("action: receive_ack | result: fail | error: Client disconnected")
            logging.info("action: agente_%d_recibio_ganadores | result: success | cantidad_ganadores: %d", agent_id, len(winners_dnis))
            protocol.close()
            return

        if code == b'\x01': # Chunk code
            self.recv_chunck(protocol, agent_id, monitor)
       
 

    def recv_chunck(self, protocol, agent_id, monitor):
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
                monitor.store_bets_thread_safe([bet])
                # utils.store_bets([bet])
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
