// Nodejs implementation for github.com/eleme/noise.

var net = require('net');
var util = require('util');

function Noise(options) {
  this.host = options.host || '0.0.0.0';
  this.port = options.port || 9000;
  this.sock = null;
  this.isPub = false;
  this.isSub = false;
}

Noise.prototype.connect = function(cb) {
  var self = this;
  this.sock = net.connect({
    host: self.host,
    port: self.port
  }, cb);
};

Noise.prototype.pub = function(name, stamp, value, cb) {
  var self = this;
  if (!this.sock)
    this.connect(function() {
      return self.sock.write("pub\n", function() {
        return self.doPub(name, stamp, value, cb);
      });
    });
  return self.doPub(name, stamp, value, cb);
};

Noise.prototype.doPub = function(name, stamp, value, cb) {
  this.sock.write(util.format("%s %d %s\n", name, stamp, value), cb);
};

Noise.prototype.sub = function(cb) {
  var self = this;
  if (!this.sock)
    this.connect(function() {
      return self.sock.write("sub\n", function() {
        return self.doSub(cb);
      });
    });
  return self.doSub(cb);
};

Noise.prototype.doSub = function(cb) {
  var self = this;
  var buf = '';
  this.sock.on('data', function(data) {
    buf += data;
    var lines = data.split('\n')
    if (data[data.length-1] === '\n') {
      buf = '';
    } else {
      buf = lines[lines.length - 1];
      lines.pop();
    }
    for (var i = 0; i < lines.length; i++) {
      var item = line.split(/\s+/);
      cb(item[0], item[1], item[2], item[3]);
    }
  });
};
