# Leader and Workers

This is a project for testing distributed workers controlled by a leader.


- one leader instance (for now) manages a group of workers
- source of truth managed by leader
- workers get updates from leader
- group of workers are selected to be active, otherwise idle
- active workers "do work"
- all workers update the leader with their health/status

## Selected Workers
Only a subset of workers are active at a time, the rest are idle.
The leader distributes configuration to tell which workers should be active.

## Failover
When the leader hasn't heard from a worker in a while it is considered unhealthy.
If that worker is a selected worker, then a new one must be selected.



## Code Flow
- Leader starts with initial configuration
- Workers all start in idle state without configuration
- Workers report health to leader and receive the configuration
- Workers that are active "do work"
- If the leader doesn't get a health update in 5 seconds, the workers is unhealthy.
- The leader selects the next idle healthy worker from the pool.
- The new configuration is received by the workers on the next health update.

## TODO
- [ ] leader removes pod that has been unhealthy for a certain time (implemented but stats not working)
- [ ] workers use context and timeouts
- [ ] workers use channels for interrupts and errors
- [ ] leader selects necessary workers in a single loop
- [ ] leader store state in redis/cache?
- [x] add prometheus metrics
