import socket
import logging
import threading
from .utils import store_bets, load_bets, has_won
from .protocol import Protocol
from .protocol import (OPCODE_INIT_CLIENT, OPCODE_BATCH_BETS, OPCODE_FINISHED_BETS, OPCODE_REQUEST_WINNERS)

class Server:
    def __init__(self, port, listen_backlog, clients_number):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._clients_threads = []
        self._clients_socket = []
        self._clients_number = clients_number
        self._agency_ready = [False] * clients_number

        # Lock to protect access to _agency_ready
        self._lock_agency_ready = threading.Lock()

        # Lock to protect access to storage from utils
        self._lock_storage = threading.Lock()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            client_sock = self.__accept_new_connection()
            client_th = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
            client_th.start()
            self._clients_threads.append(client_th)
            self._clients_socket.append(client_sock)
        

    def __handle_client_connection(self, client_sock):
        """
        Read messages from a specific client socket until the connection is closed.

        If a problem arises in the communication with the client, the
        client socket will also be closed.
        """
        protocol = Protocol(client_sock)
        try:
            while True:
                opcode = protocol.read_opcode()

                if opcode == OPCODE_INIT_CLIENT:
                    client_id = self.__handle_init_client(protocol) 

                elif opcode == OPCODE_BATCH_BETS:
                    if not self.__handle_batch_msg(protocol):
                        return
                
                elif opcode == OPCODE_FINISHED_BETS: 
                    self.__handle_finished_bets(client_id)

                elif opcode == OPCODE_REQUEST_WINNERS:
                    if self.__handle_request_winners(protocol, client_id):
                        return
                    
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()


    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c


    def stop(self):
        self._server_socket.close()
        logging.info("action: stop | result: success | detail: server socket was closed")
        for client in self._clients_socket:
            client.close()
        for thread in self._clients_threads:
            thread.join()
        logging.info(f"action: stop | result: success | detail: client socket was closed")


    # ------ Handlers for each opcode ------ #

    def __handle_batch_msg(self, protocol: Protocol) -> bool:
        """Handle receiving and storing a batch of bets. Returns False if should stop loop."""
        bets = protocol.read_batch_msg()

        if bets is None or len(bets) == 0:
            logging.warning("action: apuesta_recibida | result: fail | cantidad: 0")
            protocol.send_batch_confirmation(False)
            return False

        with self._lock_storage:
            store_bets(bets)

        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
        protocol.send_batch_confirmation(True)
        return True


    def __handle_request_winners(self, protocol: Protocol, client_id: int) -> bool:
        with self._lock_agency_ready:
            ready = all(self._agency_ready)

        if ready:
            # Todas las agencias estan listas
            
            with self._lock_storage:
                bets = load_bets()
            
            ganadores = []
            for bet in bets:
                if bet.agency == client_id and has_won(bet):
                    ganadores.append(bet.document)
            protocol.send_winners(ganadores)
            return True
        else:
            # Las agencias todavia no estan listas
            protocol.send_wait()
            return False


    def __handle_init_client(self, protocol: Protocol) -> int:
        return protocol.receive_client_id()  
    

    def __handle_finished_bets(self, client_id: int) -> None:
        with self._lock_agency_ready:
            self._agency_ready[client_id-1] = True