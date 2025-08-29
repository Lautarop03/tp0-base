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
    """
    Reads a bet message from the socket.
    Returns a Bet object.
    """
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

def send_bet_confirmation(client_sock: socket.socket, bet: Bet) -> None:
    """
    Sends the bet confirmation to the client by serializing
    all fields as strings (with uint16 length)
    and using sendall to avoid short-write.
    """
    buf = bytearray()

    # Helper to serialize strings with uint16 length + bytes
    def write_string(s: str):
        encoded = s.encode('utf-8')
        buf.extend(struct.pack('>H', len(encoded))) # big-endian uint16
        buf.extend(encoded)

    # Serialize all fields as strings
    write_string(bet.first_name)
    write_string(bet.last_name)
    write_string(bet.document)
    write_string(bet.birthdate.isoformat())
    write_string(str(bet.number))

    client_sock.sendall(buf)