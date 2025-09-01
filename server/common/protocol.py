import socket
import struct
from .utils import Bet


def read_exact(sock: socket.socket, n: int) -> bytes:
    """Reads exactly n bytes from the socket"""
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("Connection closed")
        buf += chunk
    return buf


def read_bet_msg(client_sock: socket.socket) -> Bet:
    """
    Reads a bet message from the socket.
    Returns a Bet object.
    """
    def read_string():
        # Read 1 byte of length
        len_bytes = read_exact(client_sock, 1)
        length = struct.unpack('>B', len_bytes)[0]  # uint8 big-endian
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


def read_batch_msg(client_sock: socket.socket) -> list[Bet]:
    len_batch = read_exact(client_sock, 2)
    length = struct.unpack('>H', len_batch)[0]  # uint16 big-endian

    bets = []
    for _ in range(length):
        bet = read_bet_msg(client_sock)
        bets.append(bet)
    return bets


def send_batch_confirmation(client_sock: socket.socket, success: bool, message: str = "") -> None:
    """
    Sends the batch confirmation to the client.
    Format:
        1 byte: success (0 = fail, 1 = success)
        2 bytes: message length (uint16 big-endian)
        N bytes: message (utf-8)
    """
    buf = bytearray()
    buf.append(1 if success else 0)
    encoded_msg = message.encode('utf-8')
    buf.extend(struct.pack('>H', len(encoded_msg)))
    buf.extend(encoded_msg)
    client_sock.sendall(buf)
