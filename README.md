# optional

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/optional.svg)](https://pkg.go.dev/github.com/kazhuravlev/optional)
[![License](https://img.shields.io/github/license/kazhuravlev/optional?color=blue)](https://github.com/kazhuravlev/optional/blob/main/LICENSE)
[![Build Status](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/optional)](https://goreportcard.com/report/github.com/kazhuravlev/optional)
[![codecov](https://codecov.io/gh/kazhuravlev/optional/graph/badge.svg?token=B7ILMGURZW)](https://codecov.io/gh/kazhuravlev/optional)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#utilities)

Type-safe optional values for Go - marshal/unmarshal to/from JSON, YAML, SQL and more.

## Installation

```shell
go get github.com/kazhuravlev/optional
```

Requires Go 1.23 or newer.

## Quick Start

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/kazhuravlev/optional"
)

type User struct {
	Name      string               `json:"name"`
	AvatarURL optional.Val[string] `json:"avatar_url,omitzero"` // omitzero needs Go 1.24+
}

func main() {
	// JSON with optional field
	data := []byte(`{"name": "Alice", "avatar_url": "https://example.com/avatar.jpg"}`)
	
	var user User
	json.Unmarshal(data, &user)
	
	// Check if value exists
	if avatarURL, ok := user.AvatarURL.Get(); ok {
		fmt.Printf("Avatar URL: %s\n", avatarURL)
	}
	
	// Or use a default
	url := user.AvatarURL.ValDefault("https://example.com/default.jpg")
	fmt.Printf("URL with default: %s\n", url)
}
```

## Key Features

- **Type-safe**: Generic implementation prevents runtime type errors
- **Zero value friendly**: Missing values are not mistaken for zero values
- **Multiple formats**: Works with JSON, YAML, SQL, and custom marshalers
- **Simple API**: Intuitive methods like `Get()`, `Set()`, `HasVal()`, and `Reset()`

## API Overview

```go
var opt optional.Val[string] // the zero value is an empty container

// Set a value
opt.Set("hello")

// Check and get
if val, ok := opt.Get(); ok {
    fmt.Println(val) // "hello"
}

// Get with default
val := opt.ValDefault("default") // "hello"

// Reset to empty
opt.Reset()

// Convert to pointer (nil if empty)
ptr := opt.AsPointer() // *string

// Empty? (also used by json omitzero and yaml omitempty)
opt.IsZero() // true
```

## Encoding details

### JSON

- An empty container is encoded as `null`; `null` and a missing field decode as empty.
- Decoding `null` over a container that already has a value clears it.
- Use `json:",omitzero"` (Go 1.24+) to drop empty containers from the output.
  `omitempty` has no effect on struct types, so `Val` is always emitted with it.
- A present `nil` slice or `nil` pointer is also encoded as `null` and therefore decodes
  as empty: "present but null" cannot be represented.

### YAML (gopkg.in/yaml.v3)

- An empty container is encoded as `null`; `yaml:",omitempty"` drops it.
- `null`, `~` and an empty value decode as empty when the destination is a zero struct.
  yaml.v3 never hands null nodes to custom unmarshalers, so decoding `null` over a
  container that already holds a value leaves it unchanged. Use `*optional.Val[T]` when
  "null clears the value" matters.
- yaml.v3 drops `null` elements from slices of non-pointer element types (`[]Val[T]` behaves
  like `[]string` here); `[]*Val[T]` keeps them as `nil`.

### SQL (database/sql)

- `NULL` scans as empty, even when `T` implements `sql.Scanner` itself.
- A `T` that implements `sql.Scanner` scans in place, so state prepared on the value
  beforehand stays available to it. Other values are converted like database/sql does for
  plain destinations: `int64` into any integer type, `[]byte` into `string`, and so on.
  `[]byte` (and `sql.RawBytes`) is copied, so the container never aliases a buffer owned by
  the driver.
- `Value()` asks a `T` that implements `driver.Valuer` for its value, then normalises the
  result to the `driver.Value` subset where possible (`int` and `time.Duration` become
  `int64`, named string types become `string`). Values the default converter rejects, such
  as slices, maps or structs, are handed to the driver unchanged: pgx encodes them itself,
  strict drivers report their own error.
