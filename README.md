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
CPU Request 2800m
Memory Request 13380Mi
CPU Limit 4800m
Memory Limit 20068Mi
```
Takes in account replica count on each resource.

# TODO
  - [X] Defaults support (as paramaeter as well as validation)
  - [ ] Volumes summary calculation
  - [ ] Reports generation
  - [X] Remote manifest support
  