Noise
=====

Noise is a simple daemon to detect anomalous stats, via the well-known
[3-sigma rule](https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule)
and [exponential weighted moving average](https://en.wikipedia.org/wiki/Moving_average).
Input stats stream (for example, from statsd), and noise will output the
anomalous datapoints.

Features
--------

* Automatic detection (no more thresholds).
* Follows the stats trending (via ewma).
* Minimal disk storage (detection only rely on 2 number).
* Pub/Sub implementation (clients can publish stats or subscribe anomalies).

Install
-------

    go get github.com/hit9/noise

Stats Publish
-------------

Just telnet to port 9000 and type `pub`, then send stats to noise line by line:

```bash
$ telnet 0.0.0.0 9000
pub
counter.foo 1443514447 3.14
```

Stats Subscribe
---------------

Just telnet to port 9000 and type `sub`, noise will push anomalies automatically:

```bash
$ telnet 0.0.0.0 9000
sub
counter.foo 1443515465 10.900 1.122
counter.foo 1443515495 0.900 -1.151
```

With Statsd
-----------

Install `noise-statsd` vid npm and add it to statsd's config:

```js
{
, backends: ['noise-statsd']
}
```

Detection Algorithm
-------------------

What is the [3-sigma rule](https://en.wikipedia.org/wiki/68–95–99.7_rule):
states that nearly all values (99.7%) lie within 3 standard
deviations of the mean in a normal distribution. So if a stat dosen't meet
this rule, it must be an anomaly. Describe it in pseudocode:

```python
if abs(x - avg) > 3*std:
    return True  # anomaly
```

And now we name the ratio of `abs(x - avg)` to `3 * std` as `m` (also the last
field in noise's output), then if `m > 1` the series is currently anomalous,
and the `m` large, the more serious anamlous. And more, `m > 0` shows that the
serires current trending is up, otherwise down.

How to get `avg` and `std`? We may need to store all stats on disk, each time
when new stat comes in, we should query all stats in this series from database,
and compute the `avg` and `std` via the traditional math formulas. With ewma (the
exponentially weighted moving average/standard deviation), the storage and the compution
are both not required:

```python
avgOld = avg
avg = avg * (1-f)*avg + x*f
std = sqrt((1-f)*std*std + f*(x-avgOld)*(x-avg))
```

So noise requires minimal disk storage and runs fast.

Configurations
--------------
* **port** tcp port to bind, default: `9000`
* **dbpath** leveldb database directory path, default: `noise.db`
* **factor** the ewma factor (0~1), the `factor` larger the timeliness better, default: `0.07`
* **strict** if set false, noise will use `(avg+x)/2` as new `x`, default: `true`
* **periodicity** format is `[grid, numGrids]` and we suppose that `grod*numGrids` is
  this metric's `periodicity`. default: `[480, 180]`

Net Protocol
------------

```
name timestamp value anoma '\n'
```

License
--------

MIT. (c) 2015 Chao Wang.
