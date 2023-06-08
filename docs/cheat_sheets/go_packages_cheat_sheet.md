# Go Packages Cheat Sheet

<details>
<summary>Table of Contents</summary>

 - [Basics](#basics)
 - [Imports](#imports)
 - [Exported Names](#exported-names)
 - [Creating a Package](#creating-a-package)
 - [Organizing Code with Multiple Files](#organizing-code-with-multiple-files)
 - [Package Initialization](#package-initialization)
 - [Multi-Package Project Structure](#multi-package-project-structure)
 - [Creating a Library Package](#creating-a-library-package)

</details>

This cheat sheet provides a quick reference guide for working with packages in Go. Packages are a way of organizing code in Go and they form the basis of code reuse and encapsulation.

## Basics

In Go, every file must start with a package declaration. The `main` package is a special name that tells Go to compile the package as an executable, instead of a library. 

Example:

```go
package main
```
[(back-to-top)](#go-packages-cheat-sheet)

## Imports

To use code from another package, you need to import it using the `import` keyword.

Example:

```go
package main

import (
	"fmt"
	"math"
)
```

[(back-to-top)](#go-packages-cheat-sheet)

## Exported Names

In Go, a name is exported if it begins with a capital letter. This means that it can be accessed from outside the package it is declared in.

Example:

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println(math.Pi) // Pi is exported from the math package
}
```

[(back-to-top)](#go-packages-cheat-sheet)

## Creating a Package

To create a package, simply create a new directory and add a `.go` file with the package declaration at the top.

Example:

```go
// In a file called mypackage.go
package mypackage

// Exported function
func MyFunction() {
	// Function implementation
}
```

[(back-to-top)](#go-packages-cheat-sheet)

## Organizing Code with Multiple Files

You can split a package into multiple files. As long as they are in the same directory and have the same package declaration at the top, they will be compiled as part of the same package.

Example:

```go
// In a file called file1.go
package mypackage

func Function1() {
	// Function implementation
}

// In another file called file2.go
package mypackage

func Function2() {
	// Function implementation
}
```

[(back-to-top)](#go-packages-cheat-sheet)

## Package Initialization

The `init` function can be used to perform initialization tasks. If a package has more than one `init` function (in one or more files), they are executed in the order they are presented to the compiler.

Example:

```go
package mypackage

func init() {
	// Initialization code
}
```

[(back-to-top)](#go-packages-cheat-sheet)

## Multi-package Project Structure

In larger Go programs, code is often organized into multiple packages. This helps to structure the code in a logical way and promotes code reuse and encapsulation.

Here's an example of a multi-package project:

```
/myproject
  /mypackage
    mypackage.go
  /anotherpackage
    anotherpackage.go
  main.go
```

In `myproject/mypackage/mypackage.go`:

```go
package mypackage

// Exported constant
const MyConst = "Hello, world!"

// Exported function
func MyFunction() string {
	return MyConst
}
```

In `myproject/anotherpackage/anotherpackage.go`:

```go
package anotherpackage

import (
	"fmt"
	"myproject/mypackage"
)

// Exported function
func AnotherFunction() {
	fmt.Println(mypackage.MyFunction())
}
```

In `myproject/main.go`:

```go
package main

import (
	"myproject/anotherpackage"
)

func main() {
	anotherpackage.AnotherFunction()
}
```

In this example, `mypackage` and `anotherpackage` are separate packages in the same Go module (the `myproject` module). They each have their own `.go` file and export a function that can be used by other packages. The `main.go` file imports `anotherpackage` and calls its exported function in the `main` function, which is the entry point of the program.

Note: In order for this example to work, you would need to initialize a Go module by running `go mod init myproject` in the `myproject` directory. This creates a `go.mod` file, which defines the module path (in this case, `myproject`). The module path is used as the prefix for import statements.

[(back-to-top)](#go-packages-cheat-sheet)


## Creating a Library Package

A library package in Go is a package that is intended to be imported and used by other packages, but is not itself an executable program. Library packages do not contain a `main` function.

Here's an example of a library package structure:

```
/mylib
  /mypackage
    mypackage.go
  /anotherpackage
    anotherpackage.go
```

In `mylib/mypackage/mypackage.go`:

```go
package mypackage

// Exported constant
const MyConst = "Hello, world!"

// Exported function
func MyFunction() string {
	return MyConst
}
```

In `mylib/anotherpackage/anotherpackage.go`:

```go
package anotherpackage

import (
	"fmt"
	"mylib/mypackage"
)

// Exported function
func AnotherFunction() {
	fmt.Println(mypackage.MyFunction())
}
```

In this example, `mypackage` and `anotherpackage` are separate packages in the same Go module (the `mylib` module). They each have their own `.go` file and export a function that can be used by other packages.

Note that there's no `main` function in this example, because this is a library package, not an executable program.

To use this library in another Go program, you would import it like this:

```go
import (
	"mylib/mypackage"
	"mylib/anotherpackage"
)
```

You would then be able to call `mypackage.MyFunction()` and `anotherpackage.AnotherFunction()` from your code.

Note: In order for this example to work, you would need to initialize a Go module by running `go mod init mylib` in the `mylib` directory. This creates a `go.mod` file, which defines the module path (in this case, `mylib`). The module path is used as the prefix for import statements.

[(back-to-top)](#go-packages-cheat-sheet)