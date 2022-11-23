### 结构

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

### 字段

#### outbounds

用于测试的出站标签列表。

#### providers

用于测试的[订阅](/zh/configuration/provider)标签列表。

#### fallback

如果没有符合负载均衡策略的节点，回退到此出站。

#### check

参见“健康检查字段”

#### pick

参见“节点挑选字段”

### 健康检查字段

#### interval

每个节点的健康检查间隔。不小于`10s`，默认为 `5m`。

#### sampling

对最近的多少次检查结果进行采样。大于 `0`，默认为 `10`。

#### destination

用于健康检查的链接。默认使用 `http://www.gstatic.com/generate_204`。

#### connectivity

网络连通性检查地址，默认为空。

健康检查失败，可能是由于网络不可用造成的（比如断开 WIFI 连接）。设置此项，可避免此类情况下将节点判定为失效，否则不会有此行为。

### 节点挑选字段

#### objective

负载均衡的目标。默认为 `alive`。

| 目标        | 描述                                            |
| ----------- | ----------------------------------------------- |
| `alive`     | 筛选出存活节点                                  |
| `leastload` | 筛选出较低负载的节点 （历次检查中表现更稳定的） |
| `leastping` | 筛选出较低延时的节点                            |

#### strategy

负载均衡的策略。默认为 `random`。

| 策略             | 描述                                                      |
| ---------------- | --------------------------------------------------------- |
| `random`         | 从符合目标要求的节点中随机挑选                                |
| `roundrobin`     | 从符合目标要求的节点中轮流选择                                |
| `fallback`       | 仅当前节点不符合目标要求时，重新选择                          |
| `consistenthash` | 使用同一节点处理同源站点的请求。仅当目标为 `alive` 时可用。 |

#### max_rtt

可接受的健康检查最大往返时间。 默认为 `0`，即接受任何往返时间。

#### max_fail

健康检查最大失败次。默认为 `0`，即不允许任何失败。

!!! tip "节点存活判定"

    除了确实无法连接的节点外，超过 `max_rtt` 和 `max_fail` 设置值的节点也将被判定为无效

#### expected

> 当 `objective` 为 `alive` 时，此字段不生效。

期望选出的节点数量。默认为 `0`，即自动。

#### baselines

> 当 `objective` 为 `alive` 时，此字段不生效。

选择节点的基准线，它将节点划分为不同的档位。默认为空。

- 对于 `leastload`，根据往返时间标准差划分；
- 对于 `leastping`，根据往返时间平均值划分。

!!! tip "expected 和 baselines 的作用逻辑"

    这里以策略目标 `leastping` 举例说明几种典型配置：
    
    1. 如果两者均未配置，选择出往返时间最短的一个节点。

    1. `baselines: ["500ms","700ms","900ms"]`，尝试选出往返时间在 500ms 内的节点，若没有则顺延。

    1. `expected: 3`，选出往返时间最小的 3 个节点。

    1. `expected:3, baselines =["300ms","400ms","500ms"]`，
    
        前一个配置中，假设选择了 `250ms`, `300ms`, `350ms` 的三个节点，但还有更多 `350-400ms` 的节点与被选的几乎同样优秀，我们不希望浪费它们。
    
        配置上述基准线后，要选择到 3 个节点，必须跨入 `300-400ms` 这一档位，那么这一档位内的其它节点也一同被选中。
