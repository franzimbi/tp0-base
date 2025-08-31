import socket

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
    
    def send_int(self, n):
        n_bytes = n.to_bytes(4, byteorder='little')
        self.skt.sendall(n_bytes)
        return
    
    def send_ack(self):
        self.skt.sendall(b'\x01')
        return
    
    def recv_agent_id(self):
        return self.recv_int()
    
    def send_lottery_results(self, results):
        self.send_int(len(results))
        for document in results:
            self.send_int(document)
        return

    
    def recv_code(self):
        code = self.__recv_all(1)
        if code is None:
            return None
        return code
    
    def recv_bets(self):
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
    
    def close(self):
        if self.skt is not None:
            self.skt.close()