noise-statsd
============

Noise support for statsd.

Install
-------

    npm install noise-statsd

Usage
------

Add `noise-statsd` to statsd config:

```js
, backends: ['noise-statsd']
```

Options
-------

* **noiseHost** noise host to connect, default `'0.0.0.0'`
* **noisePort** noise host to connect, default `9000`
* **noiseIgnore** patterns to ignore, default `['statsd.*']`
* **noiseTimerDataFields** timer data fields to use, default `['mean_90', 'count_ps']`

Limitation
----------

Only support for `counter_rates` and `timer_data`.
