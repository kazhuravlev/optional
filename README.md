# optional

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/optional.svg)](https://pkg.go.dev/github.com/kazhuravlev/optional)
[![License](https://img.shields.io/github/license/kazhuravlev/optional?color=blue)](https://github.com/kazhuravlev/optional/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml/badge.svg?branch=master)](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml?query=branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/optional)](https://goreportcard.com/report/github.com/kazhuravlev/optional)
[![codecov](https://codecov.io/gh/kazhuravlev/optional/graph/badge.svg?token=B7ILMGURZW)](https://codecov.io/gh/kazhuravlev/optional)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#utilities)

Type-safe optional values for Go - marshal/unmarshal to/from JSON, YAML, SQL and more.

## Installation

```shell
go get github.com/kazhuravlev/optional
```

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
	AvatarURL optional.Val[string] `json:"avatar_url,omitempty"`
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
var opt optional.Val[string]

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
```
