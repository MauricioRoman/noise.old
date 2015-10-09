// Nodejs implementation for github.com/eleme/noise.

var net = require('net');
var util = require('util');

function Noise(options) {
  options = options || {};
  this.host = options.host || '0.0.0.0';
  this.port = options.port || 9000;
  this.sock = null;
  this.isPub = false;
  this.isSub = false;
}

Noise.prototype.connect = function(cb) {
  var self = this;
  return this.sock = net.connect({
    host: self.host,
    port: self.port
  }, cb);
};

Noise.prototype.pub = function(name, stamp, value, cb) {
  var self = this;
  if (!this.sock) {
    return this.connect(function() {
      return self.sock.write("pub\n", function() {
        return self.doPub(name, stamp, value, cb);
      });
    });
  }
  return self.doPub(name, stamp, value, cb);
};

Noise.prototype.doPub = function(name, stamp, value, cb) {
  return this.sock.write(util.format("%s %d %s\n", name, stamp, value), cb);
};

Noise.prototype.sub = function(cb) {
  var self = this;
  if (!this.sock) {
    this.connect(function() {
      return self.sock.write("sub\n", function() {
        return self.doSub(cb);
      });
    });
  }
  return self.doSub(cb);
};

Noise.prototype.doSub = function(cb) {
  var self = this;
  var buf = '';
  this.sock.on('data', function(data) {
    buf += data;
    var lines = buf.split('\n')
    if (data[data.length-1] === '\n') {
      buf = '';
    } else {
      buf = lines[lines.length - 1];
      lines.pop();
    }
    for (var i = 0; i < lines.length; i++) {
      var item = lines[i].split(/\s+/);
      cb(item[0], +item[1], +item[2], +item[3]);
    }
  });
};

if (require.main === module) {
  main();
}

function main() {
  noise = new Noise();
  noise.sub(function(name, stamp, value, anoma) {
    console.log(name, stamp, value, anoma);
  });
}
