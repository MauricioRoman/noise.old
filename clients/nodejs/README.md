noise-client
============

Nodejs client implemention for github.com/eleme/noise.

Install
-------

    npm install noise-client

Usage
-----

Sub:

```js
var noise = require('noise-client').Noise({port: 9000});
noise.sub(function(name, stamp, value, anoma) {...});
```

Pub:

```js
var noise = require('noise-client').Noise({port: 9000});
noise.pub(name, stamp, value);
```
