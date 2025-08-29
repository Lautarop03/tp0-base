import socket
import struct
from .utils import Bet


def read_exact(sock: socket.socket, n: int) -> bytes:
    """Reads exactly n bytes from the socket"""
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("Connection closed before receiving all bytes")
        buf += chunk
    return buf


def read_bet_msg(client_sock: socket.socket) -> Bet:
    def read_string():
        # Read 2 bytes of length
        len_bytes = read_exact(client_sock, 2)
        length = struct.unpack('>H', len_bytes)[0]  # uint16 big-endian
        # Read the string bytes
        string_bytes = read_exact(client_sock, length)
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