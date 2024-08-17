---
sidebar_position: 5
---

# Sets

## SADD

```text title="syntax"
SADD key [element] [element ...]
```

**Time complexity**: O(N) where N is the number of elements being added.

Adds items to the set and returns the number of items added. This may not be equal to the number of items attempted to
add, because existing values are ignored.

## SCARD

```text title="syntax"
SCARD key
```

**Time complexity**: O(1)

Returns the number of items in the set

## SMEMBERS

```text title="syntax"
SMEMBERS key
```

**Time complexity**: O(N) when N equals the number of items in the set.

Returns all the items in the set.

## SDIFF

```text title="syntax"
SDIFF key [key ...]
```

**Time complexity**: O(N) when N equals total number of elements in all the given sets.

Returns the members of the set resulting from the set difference between the set denoted by the first key and all other
successive keys.

If no *other keys* are supplied, it returns the original set.

## SINTER

```text title="syntax"
SINTER key [key ...]
```

**Time complexity**: O(N) when N equals the number of items in the set.

Returns the members of the set resulting from the set intersection between the set denoted by the first key and all
other
successive keys.

If no *other keys* are supplied, it returns the original set.

## SUNION

```text title="syntax
SINTER key [key ...]
```

**Time complexity**: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.

Returns the members of the set resulting from the set union between the set denoted by the first key and all
other successive keys.

If no *other keys* are supplied, it returns the original set.

