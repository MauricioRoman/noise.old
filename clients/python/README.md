noise.py
========

Python client for github.com/eleme/noise.

Install
-------

    pip install noise.py

Usage
-----

From command line:

```bash
python -m noise
```

In Python, to make a new noise client:

```python
from noise import Noise
noise = Noise(host="0.0.0.0", port=9000)
```

To subscribe anomalies:

```python
def on_anomaly(name, stamp, value, anoma, avg_old, avg_new):
    pass
noise.sub(on_anomaly)
```

To publish stats:

```python
noise.pub(name, stamp, value)
```

Note that methd `noise.sub` will block the thread via `select`.
