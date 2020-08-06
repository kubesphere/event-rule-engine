# Event Rule Engine
[![License](http://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/KubeSphere/KubeSphere/blob/master/LICENSE)

----

## Introduction

Event Rule Engine is an rule script evaluating engine for KubeSphere event processing. It can run expression calculation on json members.

## Supported grammer

### Boolean calculation

Operation supported: and, or, not, =, !=

Variables contain json boolean member can calculate directly.

```
Examples
open and involvedObject.apiVersion="v1"
!open
open = true
open != true
```

>The bool value suport three formates, TRUE True true

### Number comparision

Operation supported: =, !=, >, <, >=, <=, in, not in

Data type supported: int, long, float, double

```
Examples
count >= 20
```

### String comparision

Operation supported: =, !=, >, <, >=, <=, in, not in, contains, not contains

```
Examples
metadata.namespace = "kube-system"
metadata.namespace in ("kube-system", "default")
metadata.namespace contains "system"
```

### Regular comparision

Operation supported: like, not like, regex, not regex

```
Examples
metadata.namespace like "ks*"
metadata.name like "redis-?"
metadata.namespace in ("kube-system", "default")
metadata.namespace contains "system"
```

>Use '*' to substitute any string, use '?' to substitute a character。
>Regexp, not regexp operation only suport standard regular syntax.

### Array comparision

Operation supported: all

```
Examples
ResponseObject.status.images[28].names[0].names = "redis"
```

>You can use '*' to substitute the index of array, this means if any element of the array match the right value, then expression holds。
>You can use "start:end" such as "1:5" to specify a subset of th array. You can omitted either the start index or the end index,
>it means the subset start at the begining of array, or end at the ending of array.

## Complex examples

```
Examples
(metadata.namespace in ("kube-system", "default") or count >= 20) and involvedObject.apiVersion contains "v1"
```