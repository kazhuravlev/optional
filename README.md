# Optional values for Go projects

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/optional.svg)](https://pkg.go.dev/github.com/kazhuravlev/optional)
[![License](https://img.shields.io/github/license/kazhuravlev/optional?color=blue)](https://github.com/kazhuravlev/optional/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml/badge.svg?branch=master)](https://github.com/kazhuravlev/optional/actions/workflows/tests.yml?query=branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/optional)](https://goreportcard.com/report/github.com/kazhuravlev/optional)
[![codecov](https://codecov.io/gh/kazhuravlev/optional/graph/badge.svg?token=B7ILMGURZW)](https://codecov.io/gh/kazhuravlev/optional)

We often need tools that help us work with optional values. For example, for optional fields in requests/responses from
APIs (JSON), optional values in YAML/JSON configs, or nullable columns in some tables. Typically, we use a pointer for
certain data types, but this does not work well for some scenarios.

This package offers a simple way to define optional fields and provides some helpers for marshalling and unmarshalling
data.

## Features

- JSON marshal/unmarshal
- YAML marshal/unmarshal
- SQL driver value/scan
- Useful functions to create and handle optional values

## Installation

```shell
go get -u github.com/kazhuravlev/optional
```

## Usage

```go
package main

import (
	"fmt"
	"encoding/json"
	"github.com/kazhuravlev/optional"
)

type User struct {
	AvatarURL optional.Val[string] `json:"avatar_url"`
	// some fields...
}

func main() {
	req := []byte(`{"avatar_url": "https://example.com/img.webp"}`)
	var user User
	if err := json.Unmarshal(req, &user); err != nil {
		panic(err)
	}

	// Check the presence of value and use it.
	if avatar, ok := user.AvatarURL.Get(); ok {
		fmt.Println("User have avatar", avatar)
	} else {
		fmt.Println("User have no avatar")
	}

	// Just check that value presented.
	if !user.AvatarURL.HasVal() {
		fmt.Println("User have no avatar")
	}

	// Use default value in case of it did not presented.
	fmt.Println("Avatar of this user is:", user.AvatarURL.ValDefault("https://example.com/default.webp"))

	// Use this api to adapt value to pointer. It will return nil when value not provided.
	avatarPtr := user.AvatarURL.AsPointer()
	fmt.Printf("Pointer to avatar: %#v\n", avatarPtr)
}
```