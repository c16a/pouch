---
sidebar_position: 3
---

# Lists

## LPUSH

```text title="syntax"
LPUSH key [element] [element ...]
```

**Time complexity**: O(N) where N is the number of elements being prepended.

Prepends elements to a list. This commands creates a new list if `key` doesn't exist yet.

## RPUSH

```text title="syntax"
RPUSH key [element] [element ...]
```

**Time complexity**: O(N) where N is the number of elements being appended.

Appends elements to a list. This commands creates a new list if `key` doesn't exist yet.

## LPOP

```text title="syntax"
LPOP key [count]
```

**Time complexity**: O(N) when N equals `count`.

Removes and returns elements from the beginning of a list, where `count` defaults to `1`.

## RPOP

```text title="syntax"
RPOP key [count]
```

**Time complexity**: O(N) when N equals `count`.

Removes and returns elements from the end of a list, where `count` defaults to `1`.

## LLEN

```text title="syntax"
LLEN key
```

**Time complexity**: O(1)

Returns the number of items in a list.

## LRANGE

```text title="syntax
LRANGE key [start] [stop]
```

Returns items from the list matching the specific indices.

The offsets `start` and `stop` are zero-based. The `start`
offset default to `0` and returns elements from the head of the list. The `end` offset defaults to `-1` and
returns elements until the tail of the list.

