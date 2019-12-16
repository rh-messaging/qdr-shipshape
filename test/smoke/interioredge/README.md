# Interior and Edge meshes

Deploys two interior router meshes (running in distinct namespaces, within the same
context). Two edge routers will connect to each interior mesh.

The network topology would be like:

```
---------------                                     ---------------
+ edge-west-1 + ---                             --- + edge-east-1 +
---------------   |                             |   ---------------
                  v                             v
           -----------------            -----------------
           + interior-west +   ----->   + interior-east +
           -----------------            -----------------
                  ^                              ^
---------------   |                             |   ---------------
+ edge-west-2 + ---                             --- + edge-east-2 +
---------------                                     ---------------
```

## Clients

A total of 72 clients (senders + receivers) will connect with the mesh above

* 2 java (cli-java) receivers will connect against each router node (12 java receivers)
* 2 java (cli-java) senders will connect against each router node (12 java senders)
* 2 python (cli-proton-python) receivers will connect against each router node (12 python receivers)
* 2 python (cli-proton-python) senders will connect against each router node (12 python senders)
* 2 nodejs (cli-rhea) receivers will connect against each router node (12 nodejs receivers)
* 2 nodejs (cli-rhea) senders will connect against each router node (12 nodejs senders)

## Tests

* Anycast and Multicast tests
* Small (1kb), medium (100kb) and large messages (1mb) to be exchanged
