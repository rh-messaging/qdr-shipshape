# Interior and Edge meshes

Deploys two interior router meshes (running in same namespace, within the same
context). 

The network topology would be like:

```
           -----------------            -----------------
           + interior-west +   ----->   + interior-east +
           -----------------            -----------------
```

## Clients

A total of 24 clients (senders + receivers) will connect with the mesh above

* 2 java (cli-java) receivers will connect against each router node (4 java receivers)
* 2 java (cli-java) senders will connect against each router node (4 java senders)
* 2 python (cli-proton-python) receivers will connect against each router node (4 python receivers)
* 2 python (cli-proton-python) senders will connect against each router node (4 python senders)
* 2 nodejs (cli-rhea) receivers will connect against each router node (4 nodejs receivers)
* 2 nodejs (cli-rhea) senders will connect against each router node (4 nodejs senders)

## Tests

* Anycast and Multicast tests
* Small (1kb), medium (100kb) and large messages (500kb) to be exchanged
