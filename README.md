# Optional values for Go projects

We are often need some tools that help us to work with optional values. For example for optional fields in
requests/responses from API (json), optional values in YAML/JSON configs or optional (nullable) columns in some tables.
Usually we use a pointer for some data types but this is works not well for some scenario.

This package provides a simple way to define those optional fields and some helpers for marshalling and unmarshalling
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
	"github.com/kazhuravlev/optional"
)
import "encoding/json"

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