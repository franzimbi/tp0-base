import socket

from common.utils import Bet

class Protocol:
    def __init__(self, socket: socket.socket):
        self.skt = socket
    
    def __recv_all(self, size):
        ''' la forma de no tener un short read'''
        buf = b''
        while len(buf) < size:
            n = self.skt.recv(size - len(buf))
            if n == 0:
                return None
            buf += n
        return buf

    def recv_int(self):
        buf = self.__recv_all(4)
        if buf is None:
            return None
        return int.from_bytes(buf, byteorder='little')
    
    def recv_string(self):
        size = self.__recv_all(1)
        if size is None:
            return None
        size = int.from_bytes(size, byteorder='little')

        string = self.__recv_all(size)
        if string is None:
            return None
        return string.decode('utf-8')
    
    def send_ack(self):
        self.skt.sendall(b'\x00')
        return
    
    def send_errorApuesta(self):
        self.skt.sendall(b'\x02')
        return
    
    def _recv_bet(self):
        nombre = self.recv_string()
        if nombre is None:
            raise OSError("Client disconnected")
        apellido = self.recv_string()
        if apellido is None:
            raise OSError("Client disconnected")
        documento = self.recv_int()
        if documento is None:
            raise OSError("Client disconnected")
        nacimiento = self.recv_string()
        if nacimiento is None:
            raise OSError("Client disconnected")
        numero = self.recv_int()
        if numero is None:
            raise OSError("Client disconnected")
        
        return (nombre, apellido, documento, nacimiento, numero)
    
    def recv_agency_id(self):
        return self.recv_int()
    
    def continue_recv_chuncks(self):
        buf =  self.skt.__recv_all(1)
        if buf is None:
            raise OSError("Client disconnected")
        if buf == b'\x00':
            return True
        else:
            return False

    def recv_bets(self, agency_id):
        count = self.recv_int()
        if count is None:
            raise OSError("Client disconnected")  
        bets = []
        for _ in range(count):
            bet = self._recv_bet()
            bets.append(Bet(agency_id, bet[0], bet[1], str(bet[2]), bet[3], str(bet[4])))
        return bets
    
    def close(self):
        if self.skt is not None:
            self.skt.close()