import socket
import logging
from .utils import store_bets, load_bets, has_won
from .protocol import Protocol
from .protocol import (OPCODE_INIT_CLIENT, OPCODE_BATCH_BETS, OPCODE_FINISHED_BETS, OPCODE_REQUEST_WINNERS)

class Server:
    def __init__(self, port, listen_backlog, clients_number):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket = None
        self._clients_number = clients_number
        self._agency_ready = [False] * clients_number
        self.protocol = Protocol(self._client_socket)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            client_sock = self.__accept_new_connection()
            self._client_socket = client_sock
            self.__handle_client_connection(client_sock)


    def __handle_client_connection(self, client_sock):
        """
        Read messages from a specific client socket until the connection is closed.

        If a problem arises in the communication with the client, the
        client socket will also be closed.
        """
        try:
            self.protocol.socket = client_sock

            while True:
                opcode = self.protocol.read_opcode()

                if opcode == OPCODE_INIT_CLIENT:
                    client_id = self.__handle_init_client() 

                elif opcode == OPCODE_BATCH_BETS:
                    if not self.__handle_batch_msg():
                        return
                
                elif opcode == OPCODE_FINISHED_BETS: 
                    self.__handle_finished_bets(client_id)
                    return

                elif opcode == OPCODE_REQUEST_WINNERS:
                    self.__handle_request_winners(client_id)
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
        self._client_socket.close()
        logging.info(f"action: stop | result: success | detail: client socket was closed")


    # ------ Handlers for each opcode ------ #

    def __handle_batch_msg(self) -> bool:
        """Handle receiving and storing a batch of bets. Returns False if should stop loop."""
        bets = self.protocol.read_batch_msg()

        if bets is None or len(bets) == 0:
            logging.warning("action: apuesta_recibida | result: fail | cantidad: 0")
            self.protocol.send_batch_confirmation(False, "No se recibieron apuestas")
            return False

        store_bets(bets)
        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
        self.protocol.send_batch_confirmation(True, "Apuestas recibidas correctamente")  # TODO: cambiar por un opcode
        return True


    def __handle_request_winners(self, client_id: int) -> None:
        if all(self._agency_ready):
            # Todas las agencias estan listas
            bets = load_bets()
            ganadores = []
            for bet in bets:
                if bet.agency == client_id and has_won(bet):
                    ganadores.append(bet.document)
            self.protocol.send_winners(ganadores)
        else:
            # Las agencias todavia no estan listas
            self.protocol.send_wait()


    def __handle_init_client(self) -> int:
        return self.protocol.receive_client_id()  
    

    def __handle_finished_bets(self, client_id: int) -> None:
        self._agency_ready[client_id-1] = True