etcd peer to peer communication workflow
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
September 10, 2025
</span>

# Introduction
I ended up spending a whole day on this diagram just to answer a question on network communication metrics and clear up my confusion about etcd peer-to-peer communication.
Snapshot sender isn't included, but it's also using the pipeline transport.


![etcd peer to peer communication workflow](images/rafthttp.png)