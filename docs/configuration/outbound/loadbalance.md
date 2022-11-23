### Structure

```json
{
  "type": "loadbalance",
  "tag": "loadbalance",
  "outbounds": [
    "proxy-a",
    "proxy-b",
    "proxy-c"
  ],
  "providers": [
    "provider-a",
    "provider-b",
  ],
  "fallback": "block",
  "check": {
    "interval": "5m",
    "sampling": 10,
    "destination": "http://www.gstatic.com/generate_204",
    "connectivity": "http://connectivitycheck.platform.hicloud.com/generate_204"
  },
  "pick": {
    "objective": "leastload",
    "strategy": "random",
    "max_fail": 0,
    "max_rtt": "1000ms",
    "expected": 3,
    "baselines": [
      "50ms",
      "100ms",
      "150ms",
      "200ms",
      "250ms",
      "350ms"
    ]
  }
}
```

### Fields

#### outbounds

List of outbound tags to test.

#### outbounds

List of [Provider](/configuration/provider) tags to test.

#### fallback

The fallback outbound tag. if no outbound matches the policy, the fallback outbound will be used.

#### check

See "Check Fields"

#### pick

See "Pick Fields"

### Check Fields

#### interval

The interval of health check for each node. Must be greater than `10s`, default is `5m`.

#### sampling

The number of recent health check results to sample. Must be greater than `0`, default is `10`.

#### destination

The destination URL for health check. Default is `http://www.gstatic.com/generate_204`.

#### connectivity

The destination of connectivity check. Default is empty. 

If a health check fails, it may be caused by network unavailability (e.g. disconnected from WIFI). Set this field to avoid the node being judged to be invalid under such circumstances, or this behavior will not occur.

### Pick Fields

#### objective

The objective of load balancing. Default is `alive`.

| Objective   | Description                        |
| ----------- | ---------------------------------- |
| `alive`     | Pick from alive nodes              |
| `leastload` | Pick from nodes with lower load    |
| `leastping` | Pick from nodes with lower latency |

#### strategy

The strategy of load balancing. Default is `random`.

| Strategy         | Description                                            |
| ---------------- | ------------------------------------------------------ |
| `random`         | Randomly pick a node from ones who match the objective |
| `roundrobin`     | Rotate from nodes who match the objective              |
| `fallback`       | Re-pick only if current node not matches the objective |
| `consistenthash` | Use same node for requests to same origin targets.     |

Note: `consistenthash` is available only when the objective is `alive`

#### max_rtt

The maximum round-trip time of health check that is acceptable. Default is `0`, which accepts any round-trip time.

#### max_fail

The maximum number of health check failures, default is `0`, i.e. no failures allowed.

!!! tip "Dead Nodes"

    In addition to nodes that are unreachable, nodes that exceed the values set by `max_rtt` and `max_fail` will also be judged as dead.

#### expected

> This field has no effect, if `objective` is `alive`.

The expected number of outbound to be selected. The default value is 0, i.e. the number is automatically determined.


#### baselines

> This field has no effect, if `objective` is `alive`.

The baselines divide the nodes into different ranges. The default value is empty.

- for the `leastload`, it divides according to the standard deviation of RTTs;
-  for the `leastping`, it divides according to the average of RTTs.

!!! tip  "How does expected and baselines works"

    Suppose the `objective` is `leastping`, following examples shows some typical configurations:

    1. Select one node with least RTT, if both not configured.

    1. `baselines: ["300ms","500ms","700ms"]`,try to select nodes with RTTs within 300ms, try next base line if no matches.

    1. `expected: 3`, select 3 nodes with the least RTTs in recent checks.

    1. `expected:3, baselines =["300ms","400ms","500ms"]`,
    
        In the previous example, suppose 3 nodes of `250ms`, `300ms`, `350ms` are selected, but there are more nodes of `350-400ms`, which are almost as good as the selected ones, we do not hope to waste them.
    
        With the above baselines, to select 3 nodes, it must step into the `300-400ms` range, then other nodes in this range are also selected.
