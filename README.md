# Helm resource plugin
Calculates summary resource usage of the chart.
# Installateion 
```
    helm plugin install https://github.com/m-pavel/helm-resource
```
# Usage 
```
    helm resource sum .
```
Example output
```
CPU Limit 3600m + 1200m (Jobs) = 4800m
Memory Limit 14972Mi + 4996Mi (Jobs) = 19968Mi
CPU Request 2050m + 750m (Jobs) = 2800m
Memory Request 10580Mi + 2700Mi (Jobs) = 13280Mi
```
Takes in account replica count on each resource.

# TODO
  - [X] Defaults support (as paramaeter as well as validation)
  - [ ] Volumes summary calculation
  - [ ] Reports generation
  - [X] Remote manifest support
  - [X] Validate require requirements fits quota
  