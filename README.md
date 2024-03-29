Noise
=====

Note: Not Production Ready.

Noise is a simple daemon to detect anomalous stats, via the well-known
[3-sigma rule (68-95-99.7 rule)](https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule)
and [exponential weighted moving average (ewma)](https://en.wikipedia.org/wiki/Moving_average).
It can detect anomalies in time series **without any preconfigured thresholds**. Just input
stats stream (i.e. from [statsd](https://github.com/etsy/statsd)), noise will filter out
the anomalies and broadcast them to subscribers.

Use Case
--------

We can use noise to monitor website/rpc interfaces, including api called frequency,
api response time(time cost per call) and exceptions count. For example, we have an api
named `get_name`, this api's response time (in ms) is reported to noise from statsd every
10 seconds:

    51, 53, 49, 48, 45, 50, 51, 52, 55, 56, .., 300

Noise will catch the latest stat `300` and report it as an anomaly.

Why don't we just set a fixed threshold instead (i.e. 200ms)? This may also work but we may
have a lot of apis to monitor, some are fast (~10ms) and some are slow (~1000ms), it is hard
to set a good threshold for each one, and also hard to set an appropriate global threshold
for all. Noise sloves this via 3-sigma and ewma, it gives dynamic thresholds for each metric
("learned" from history data). We don't have to set a threshold for each metric, it will find
the "thresholds" automatically.

Features
--------

* Automatic anomalies detection, no more thresholds.
* Minimal disk storage, each metric relies on only 2 numbers.
* Pub/Sub client implementation, publish stats and subscribe anomalies.
* Following the stats trending (via ewma).

Installation
------------

    go get github.com/eleme/noise

Command Line Usage
------------------

    $ ./noise -detector -config ./config.json 
    2015/09/29 17:36:23 reading config from ./config.json..
    2015/09/29 17:36:23 listening on 0.0.0.0:9000..

Publish Stats
-------------

Just telnet to port 9000 and type `pub`, then send stats to noise line by line:

    $ telnet 0.0.0.0 9000
    pub
    counter.foo 1443514447 3.14

Input: `name timestamp value`

Subscribe Anomalies
--------------------

Just telnet to port 9000 and type `sub`, noise will push anomalies automatically:

    $ telnet 0.0.0.0 9000
    sub
    counter.foo 1443515465 10.900 1.122 10.8 19.2
    counter.foo 1443515495 0.900 -1.151 10.7 7.9

Output: `name timestamp value anomalous_serverity average_old average_new`

Publish from Statsd
-------------------

Install `noise-statsd` via npm and add it to statsd's backends list in config:

    {
    , backends: ['noise-statsd']
    }

Detection Algorithm
-------------------

The core algorithm is the [3-sigma rule](https://en.wikipedia.org/wiki/68–95–99.7_rule):
states that nearly all values (99.7%) lie within 3 standard
deviations of the mean in a normal distribution. So if a stat dosen't meet
this rule, it must be an anomaly. To describe it in pseudocode:

```python
if abs(x-avg) > 3*std:
    return True  # anomaly
```

And now we name the ratio of `abs(x-avg)` to `3*std` as `m`:

```python
m = abs(x-avg) / (3.0*std)
```

`m` is also the last field in noise's output (when you subscribe anomalies
from it). If `abs(m)>1`, that means the series is currently anomalous, and the 
`abs(m)` large, the more serious anamlous. And more, `m > 0` shows that the
serires current trending is up, otherwise down.

How to get `avg` and `std`? We may need to store all stats on disk, each time
when new stat comes in, we should query all stats in this series from database,
and compute the `avg` and `std` via the traditional math formulas. With ewma (the
exponentially weighted moving average/standard deviation), the storage and the compution
are both not required (the `f` in following code is a float between 0 and 1):

```python
avgOld = avg
avg = avg * (1-f)*avg + x*f
std = sqrt((1-f)*std*std + f*(x-avgOld)*(x-avg))
```

The above recursive formulas make `avg` and `std` following stats trending. By this way,
noise just requires 2 numbers (the moving `avg` and `std`) to store on disk, and the
compution is simple, fast.

Configurations
--------------

* **debug** debug mode, default: `false`
* **port** tcp port to bind, default: `9000`
* **dbfile** leveldb database directory path, default: `stats.db`
* **factor** the ewma factor (0~1), the `factor` larger the timeliness better, default: `0.07`
* **strict** if set false, noise will use `(avg+x)/2` as new `x`, default: `true`
* **periodicity** its format is `[grid, numGrids]` and we suppose that `grod*numGrids` is
  this metric's `periodicity`. default: `[480, 180]`
* **whitelist** wildcards list to allow stats passing, default: `["*"]`
* **blacklist** wildcards list to disallow stats passing, default: `["statsd.*"]`

Net Protocol
------------

```
name timestamp value anoma '\n'
```

License
--------

MIT. (c) 2015 Chao Wang, Eleme, Inc. <hit9@icloud.com>
