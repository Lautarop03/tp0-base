import socket
import struct # TODO: CAMBIAR ESTO NO SE PUEDE USAR
from .utils import Bet


OPCODE_INIT_CLIENT = 0x01
OPCODE_BATCH_BETS = 0x02
OPCODE_FINISHED_BETS = 0x04
OPCODE_REQUEST_WINNERS = 0x05
OPCODE_WAIT = 0x06
OPCODE_WINNERS = 0x07


def read_exact(sock: socket.socket, n: int) -> bytes:
    """Reads exactly n bytes from the socket"""
    buf = b''
    while len(buf) < n:
        chunk = sock.recv(n - len(buf))
        if not chunk:
            raise ConnectionError("Connection closed")
        buf += chunk
    return buf


def read_opcode(client_sock: socket.socket):
    """
    Reads the opcode from the client socket.
    """
    buf = read_exact(client_sock, 1)
    if not buf:
        raise ConnectionError("Connection closed") # TODO: ??
    return buf[0]


def read_bet_msg(client_sock: socket.socket) -> Bet:
    """
    Reads a bet message from the socket.
    Returns a Bet object.
    """
    def read_string():
        # Read 1 byte of length
        len_bytes = read_exact(client_sock, 1)
        length = struct.unpack('>B', len_bytes)[0]  # uint8 big-endian # TODO: CAMBIAR ESTO NO SE PUEDE USAR
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
    length = struct.unpack('>H', len_batch)[0]  # uint16 big-endian # TODO: CAMBIAR ESTO NO SE PUEDE USAR

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


def receive_client_id(client_sock):
    """
    Receives a message with one byte indicating the length of the client ID, followed by the client ID itself.
    """
    msg = read_exact(client_sock, 1)
    client_id_length = msg[0]
    client_id_bytes = read_exact(client_sock, client_id_length)
    client_id_str = client_id_bytes.decode('utf-8')
    client_id = int(client_id_str)
    return client_id


def send_wait(client_sock: socket.socket):
    """
    Sends a message to the client indicating that not all agencies have finished yet.
    Format:
        1 byte: opcode (0x06)
    """
    buf = bytearray()
    buf.append(OPCODE_WAIT)
    client_sock.sendall(buf)


def send_winners(client_sock: socket.socket, winners: list[str]):
    """
    Sends the winners to the client.
    Format:
        1 byte: opcode (0x07)
        2 bytes: number of winners (uint16 big-endian)
        For each winner:
            1 byte: length of document (uint8)
            N bytes: document (utf-8)
    """
    buf = bytearray()
    buf.append(OPCODE_WINNERS)
    buf.extend(struct.pack('>H', len(winners)))
    for document in winners:
        encoded_doc = document.encode('utf-8')
        buf.append(len(encoded_doc))
        buf.extend(encoded_doc)
    client_sock.sendall(buf)
