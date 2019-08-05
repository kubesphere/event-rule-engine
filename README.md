# Event Rule Engine
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)

----

## Introduction

Event Rule Engine is an rule script evaluating engine for KubeSphere event processing. It can run expression calculation on json members.

## Supported grammer

### Boolean calculation

Operation supported: and or not

Variables contain json boolean member can calculate directly.

```
Examples
open and involvedObject.apiVersion="v1"
```

### Number comparision

Operation supported: =  !=  >  <  >=  <=

Data type supported: int float

```
Examples
count >= 20
```

### String comparision

Operation supported: =  in contains

```
Examples
metadata.namespace = "kube-system"
metadata.namespace in ("kube-system", "default")
metadata.namespace contains "system"
```

## Complex examples

```
Examples
(metadata.namespace in ("kube-system", "default") or count >= 20) and (not open or involvedObject.apiVersion contains "v1")
```