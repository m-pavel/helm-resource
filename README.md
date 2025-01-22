# Helm resource plugin
Calculates summary resource usage of the chart.
# Installation 
```
    helm plugin install https://github.com/m-pavel/helm-resource
```
# Usage 
## Chart summary
Calculate chart resource requirements
```
    helm resource sum .
```
Calculate deployed chart resource requirements
```
    helm resource sum <deployment-name> --remote
```

Example output
```
CPU Limit 3600m + 1200m (Jobs) = 4800m
Memory Limit 14972Mi + 4996Mi (Jobs) = 19968Mi
CPU Request 2050m + 750m (Jobs) = 2800m
Memory Request 10580Mi + 2700Mi (Jobs) = 13280Mi
```
Takes in account replica count on each resource.

## Quota validation
Calculate chart resource requirements and check it fits k8s quota
```
    helm resource check .
```
Example output
```
+----------------+---------------+---------------+---------------+---------------+---------------+--------------+
|                | Static wrkld  | Jobs          | Sum           | Quota         | Status ststic |Status sum    |
+----------------+---------------+---------------+---------------+---------------+---------------+--------------+
|CPU Limit       |         3600m |         1200m |         4800m |             8 |          true |         true |
|Memory Limit    |       14972Mi |        4996Mi |       19968Mi |          20Gi |          true |         true |
|CPU Request     |         2050m |          750m |         2800m |             8 |          true |         true |
|Memory Request  |       10580Mi |        2700Mi |       13280Mi |          20Gi |          true |         true |
|Storage Request |          18Gi |             0 |          18Gi |          50Gi |          true |         true |
+----------------+---------------+---------------+---------------+---------------+---------------+--------------+
|     configmaps |               |               |             1 |           100 |               |         true |
|        secrets |               |               |             1 |           100 |               |         true |
|       services |               |               |            14 |           100 |               |         true |
|persistentvolum |               |               |             6 |            10 |               |         true |
+----------------+---------------+---------------+---------------+---------------+---------------+--------------+
```
# TODO
  - [X] Defaults support (as paramaeter as well as validation)
  - [X] Volumes summary calculation
  - [ ] Reports generation
  - [X] Remote manifest support
  - [X] Validate require requirements fits quota
  