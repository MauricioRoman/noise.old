/**
 * Copyright (c) 2015, Chao Wang <hit9@ele.me>
 * All rights reserved by Eleme, Inc.
 *
 *  Statsd backend to work as a noise pub client.
 *  Optional configs:
 *
 *   noisHost, default: '0.0.0.0'
 *   noisePort, default: 9000
 *   noiseIgnores, default: ['statsd.*']
 *   noiseTimerDataFields, default: ['mean_90', 'count_ps']
 *
 * Metric types supported: `counter_rates` & `timer_data`.
 */

'use strict';

var net = require('net');
var minimatch = require('minimatch');

var config;
var debug;
var logger;

var makers = {
  'counter_rates': function(k, v, t) {
    return [['counter.' + k, t, v]];
  },
  'timer_data': function (k, d, t) {
    var fields = config.noiseTimerDataFields ||
      ['mean_90', 'count_ps'], datapoints = [];
    for (var i = 0; i < fields.length; i++) {
      var field = fields[i];
      var name = ['timer', field, k].join('.');
      var v = d[field];
      datapoints.push([name, t, v]);
    }
    return datapoints;
  }
};

function Noise() {
}

Noise.prototype.connect = function(cb) {
  var self = this;
  this.conn = net.connect({
    host: config.noiseHost || '0.0.0.0',
    port: config.noisePort || 9000
  }, function() {
    this.write('pub\n', function() {
      if (debug) {
        logger.log('noise connected, action: pub');
      }
      if (cb)
        cb();
    });
  })
  .on('error', function(err) {
    if (debug) {
      logger.log('noise socket error: ' + err.message);
    }
    self.conn.destroy();
    self.conn = undefined;
  });
  return this;
};

Noise.prototype.match = function(key) {
  var ignores = config.noiseIgnores || ['statsd.*'];
  for (var i = 0; i < ignores.length; i++)
    if (minimatch(key, ignores[i]))
      return true;
  return false;
}

Noise.prototype.send = function(buf, cb) {
  var self = this;
  if (!this.conn)
    return this.connect(function() {
      self.conn.write(buf, cb);
    });
  this.conn.write(buf, cb);
};

Noise.prototype.flush = function(time, data) {
  var list = [];
  var types = Object.keys(makers);
  for (var i = 0; i < types.length; i++) {
    var type = types[i];
    var dict = data[type];
    for (var key in dict) {
      if (!this.match(key)) {
        var val = dict[key];
        var maker = makers[type];
        [].push.apply(list, maker(key, val, time));
      }
    }
  }
  var length = list.length;
  if (length > 0) {
    var buf = '';
    for (i = 0; i < list.length; i++)
      buf += list[i].join(' ') + '\n';
      this.send(buf, function() {
        if (debug) {
          var msg = 'sent to noise: ' + JSON.stringify(list[0]);
          if (length > 1)
            msg += ', (' + (length - 1) + ' more..)';
          logger.log(msg);
        }
      });
  }
};

exports.init = function(uptime, _config, events, _logger) {
  var noise;
  logger = _logger || console;
  debug = _config.debug;
  config = _config || {};
  noise = new Noise();
  events.on('flush', function(time, data) {
    noise.flush(time, data);
  });
  noise.connect();
  return true;
};
