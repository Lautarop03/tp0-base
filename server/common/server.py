import socket
import struct
import logging
from .utils import Bet, store_bets

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
            bet = recibir_bet_msg(client_sock)
            store_bets([bet])
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')

            # TODO: Modify the send to avoid short-writes
            client_sock.send("{}\n".format("apuesta_almacenada").encode('utf-8')) # TODO: ver que devolver
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
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

def recv_exact(sock: socket.socket, n: int) -> bytes:
    """Lee exactamente n bytes del socket"""
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("Conexión cerrada antes de recibir todos los bytes")
        buf += chunk
    return buf


def recibir_bet_msg(client_sock: socket.socket) -> Bet:
    def read_string():
        # Leer 2 bytes de longitud
        len_bytes = recv_exact(client_sock, 2)
        length = struct.unpack('>H', len_bytes)[0]  # uint16 big-endian
        # Leer los bytes del string
        string_bytes = recv_exact(client_sock, length)
        return string_bytes.decode('utf-8')

    client_id = read_string()
    nombre = read_string()
    apellido = read_string()
    documento = read_string()
    nacimiento = read_string()
    numero = read_string()

    return Bet(
        agency= client_id, 
        first_name= nombre,
        last_name= apellido,
        document= documento,
        birthdate= nacimiento,
        number= numero )