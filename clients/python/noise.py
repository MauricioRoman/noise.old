# coding=utf8

"""Python client for github.com/eleme/noise::

    pip install noise

command line usage::

    python -m noise.py

library usage::

    from noise import Noise
    noise = Noise(host="0.0.0.0", port=9000)

sub usage::

    def on_anomaly(name, stamp, value, anoma):
        pass
    noise.sub(on_anomaly)

pub usage::

    noise.pub(name, stamp, value)

Note that meth:`noise.sub` will block the thread via `select`.
"""

import socket
import select


class Noise(object):

    def __init__(self, host="0.0.0.0", port=9000):
        self.host = host
        self.port = port
        self.sock = None
        self.is_pub = None
        self.is_sub = None

    def connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((self.host, self.port))

    def close():
        self.sock.close()

    def pub(self, name, stamp, value):
        if self.is_sub:
            raise RuntimeError("Cannot pub in sub mode")
        if self.sock is None:
            self.connect()
        if self.is_pub is None:
            self.sock.send("pub\n")
            self.is_pub = True
        line = "%s %d %.5f\n" % (name, stamp, value)
        return self.sock.send(line)

    def sub(self, on_anomaly):
        if self.is_pub:
            raise RuntimeError("Cannot sub in pub mode")
        if self.sock is None:
            self.connect()
        if self.is_sub is None:
            self.sock.send("sub\n")
            self.is_sub = True
        buf = ''
        while 1:
            rlist, wlist, xlist = select.select([self.sock], [], [])
            if not rlist:
                continue
            while 1:
                temp = self.sock.recv(1024)
                if not temp:
                    break
                buf += temp
                lines = buf.splitlines()
                if buf.endswith('\n'):
                    buf = ''
                else:
                    buf = lines[-1]
                    lines = lines[:-1]
                for line in lines:
                    args = line.split()
                    on_anomaly(*args)


if __name__ == '__main__':
    def on_anomaly(name, stamp, value, anoma):
        print name, stamp, value, anoma
    noise = Noise()
    noise.sub(on_anomaly)
