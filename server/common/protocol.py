import socket
from .utils import Bet


OPCODE_INIT_CLIENT = 0x01
OPCODE_BATCH_BETS = 0x02
OPCODE_BATCH_CONFIRMATION = 0x03
OPCODE_FINISHED_BETS = 0x04
OPCODE_REQUEST_WINNERS = 0x05
OPCODE_WAIT = 0x06
OPCODE_WINNERS = 0x07

class Protocol:
    def __init__(self, socket: socket.socket):
        self.socket = socket

    def read_uint8(self) -> int:
        return int.from_bytes(self.read_exact(1), "big")


    def read_uint16(self) -> int:
        return int.from_bytes(self.read_exact(2), "big")


    def write_uint16(self, value: int) -> bytes:
        return value.to_bytes(2, "big")


    def read_exact(self, n: int) -> bytes:
        """Reads exactly n bytes from the socket"""
        buf = b''
        while len(buf) < n:
            chunk = self.socket.recv(n - len(buf))
            if not chunk:
                raise ConnectionError("Connection closed")
            buf += chunk
        return buf


    def read_opcode(self) -> int:
        """
        Reads the opcode from the client socket.
        """
        buf = self.read_exact(1)
        return buf[0]

    def read_string(self) -> str:
        # Read 1 byte of length
        length = self.read_uint8()
        # Read the string bytes
        string_bytes = self.read_exact(length)
        return string_bytes.decode('utf-8')

    def read_bet_msg(self) -> Bet:
        """
        Reads a bet message from the socket.
        Returns a Bet object.
        """

        client_id = self.read_string()
        nombre = self.read_string()
        apellido = self.read_string()
        documento = self.read_string()
        nacimiento = self.read_string()
        numero = self.read_string()

        return Bet(
            agency= client_id, 
            first_name= nombre,
            last_name= apellido,
            document= documento,
            birthdate= nacimiento,
            number= numero )


    def read_batch_msg(self) -> list[Bet]:
        length = self.read_uint16()

        bets = []
        for _ in range(length):
            bet = self.read_bet_msg()
            bets.append(bet)
        return bets


    def send_batch_confirmation(self, success: bool) -> None:
        """
        Sends the batch confirmation to the client without a message or length.
        Format:
            1 byte: opcode = 0x03
            1 byte: success (0 or 1)
        """
        buf = bytearray()
        buf.append(OPCODE_BATCH_CONFIRMATION)
        buf.append(1 if success else 0)
        self.socket.sendall(buf)


    def receive_client_id(self) -> int:
        """
        Receives a message with one byte indicating the length of the client ID, followed by the client ID itself.
        """
        client_id_str = self.read_string()
        client_id = int(client_id_str)
        return client_id


    def send_wait(self) -> None:
        """
        Sends a message to the client indicating that not all agencies have finished yet.
        Format:
            1 byte: opcode (0x06)
        """
        buf = bytearray()
        buf.append(OPCODE_WAIT)
        self.socket.sendall(buf)


    def send_winners(self, winners: list[str]) -> None:
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
        buf.extend(self.write_uint16(len(winners)))

        for document in winners:
            encoded_doc = document.encode('utf-8')
            buf.append(len(encoded_doc))
            buf.extend(encoded_doc)
        self.socket.sendall(buf)
