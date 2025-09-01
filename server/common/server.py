import socket
import logging
from .utils import store_bets
from .protocol import read_batch_msg, send_batch_confirmation

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_socket: socket.socket = None 

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
            self._client_socket = None

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bets = read_batch_msg(client_sock)
            
            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

            send_batch_confirmation(client_sock, True, "Apuestas recibidas correctamente")
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: ${len(bets)}")
            send_batch_confirmation(client_sock, False, "Error al recibir las apuestas")
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
        if self._client_socket:
            self._client_socket.close()
            logging.info("action: stop | result: success | detail: client socket was closed")
