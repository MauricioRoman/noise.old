// Statsd backend for github.com/eleme/noise.
'use strict';

var net = require('net');
var util = require('util');
var minimatch = require('minimatch');

var config;
var debug;
var logger;

var makers =  {
  'counter_rates': dataFromCounterRates,
  'timer_data': dataFromTimerData,
};

function dataFromCounterRates(key, val, timeStamp) {
  var name = util.format('counter.%s', key)
  return [[name, timeStamp, val]];
}

function dataFromTimerData(key, dict, timeStamp) {
  var fields = config.noiseTimerDataFields || ['mean_90', 'count_ps'];
  var data = [];

  for (var i = 0; i < fields.length; i++) {
      var field = fields[i];
      var name = util.format('timer.%s.%s', field, key);
      data.push([name, timeStamp, dict[field]]);
  }
  return data
}

function Noise(options) {
  options = options || {};
  this.host = options.host || '0.0.0.0';
  this.port = options.port || 9000;
}

Noise.prototype.connect = function(cb) {
  var self = this;
  var options = {
    host: this.host,
    port: this.port,
  };
  this.conn = net.connect(options, function() {
    this.write('pub\n', function() {
      if (debug)
        logger.log('noise connected, action: pub');
      if (cb)
        return cb();
    });
  })
  .on('error', function(err) {
    if (debug)
      logger.log('noise socket error: ' + err.message);
    self.conn.destroy();
    self.conn = undefined;
  });
  return this;
};

Noise.prototype.match = function(key) {
  var patterns = config.noiseIgnores || ['statsd.*'];
  for (var i = 0; i < patterns.length; i++)
    if (minimatch(key, patterns[i]))
      return true;
  return false;
}

Noise.prototype.send = function(buf, cb) {
  var self = this;
  if (!this.conn)
    return this.connect(function() {
      return self.conn.write(buf, cb);
    });
  return this.conn.write(buf, cb);
};

Noise.prototype.flush = function(timeStamp, data) {
  var list = [];
  var types = Object.keys(makers);

  for (var i = 0; i < types.length; i++) {
    var type = types[i];
    var dict = data[type];
    for (var key in dict) {
      if (!this.match(key)) {
        var val = dict[key];
        var maker = makers[type];
        if (maker)
          [].push.apply(list, maker(key, val, timeStamp));
      }
    }
  }

  if (list.length > 0) {
    var buf = '';
    for (i = 0; i < list.length; i++)
      buf += list[i].join(' ') + '\n';
      this.send(buf, function() {
        if (debug) {
          var msg = util.format("sent to noise: %s", JSON.stringify(list[0]));
          if (list.length > 1)
            msg = util.format("%s, (%d more..)", msg, list.length - 1)
          logger.log(msg);
        }
      });
  }
};

exports.init = function(uptime, _config, events, _logger) {
  logger = _logger || console;
  debug = _config.debug;
  config = _config || {};

  var noise = new Noise({
    host: config.noiseHost,
    port: config.noisePort,
  });
  events.on('flush', function(timeStamp, data) {
    noise.flush(timeStamp, data);
  });
  noise.connect();
  return true;
};
