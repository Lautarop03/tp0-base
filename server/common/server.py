import socket
import logging
from .utils import store_bets, load_bets, has_won
from .protocol import read_batch_msg, send_batch_confirmation, read_opcode, send_wait, send_winners, receive_client_id
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
            while True:
                opcode = read_opcode(client_sock)

                if opcode == OPCODE_INIT_CLIENT:
                    client_id = receive_client_id(client_sock)   

                if opcode == OPCODE_BATCH_BETS:

                    bets = read_batch_msg(client_sock)

                    if bets is None:
                        break

                    if len(bets) == 0:
                        logging.warning("action: apuesta_recibida | result: fail | cantidad: 0")
                        send_batch_confirmation(client_sock, False, "No se recibieron apuestas")
                        break

                    store_bets(bets)
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
                    send_batch_confirmation(client_sock, True, "Apuestas recibidas correctamente") # TODO: cambiar por un opcode
                
                elif opcode == OPCODE_FINISHED_BETS: 
                    self._agency_ready[client_id-1] = True
                    return

                elif opcode == OPCODE_REQUEST_WINNERS:

                    if all(self._agency_ready):
                        # Todas las agencias estan listas
                        bets = load_bets()
                        ganadores = []

                        for bet in bets:
                            if bet.agency == client_id and has_won(bet):
                                ganadores.append(bet.document)

                        send_winners(client_sock, ganadores)
                    else:
                        # Las agencias todavia no estan listas
                        send_wait(client_sock)

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
