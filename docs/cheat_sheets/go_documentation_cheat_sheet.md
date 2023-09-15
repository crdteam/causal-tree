# Go Documentation Commentaries Cheat Sheet

<details>
<summary>Table of Contents</summary>

 - [Basics](#basics)
 - [Documenting Packages](#documenting-packages)
 - [Documenting Functions](#documenting-functions)
 - [Documenting Constants, Variables, and Types](#documenting-constants-variables-and-types)
 - [Adding Examples](#adding-examples)
 - [Using go doc and godoc](#using-go-doc-and-godoc)

</details>

Go has a robust system for documenting code that promotes best practices for clarity, readability, and maintainability. This cheat sheet highlights these best practices.

## Basics

In Go, documentation comments are written as regular comments that are placed directly above the declaration of the item they are documenting. They should start with the name of the item being documented and be written in complete sentences.

Multi-line comments can be used to provide more detailed documentation. Pre-formatted text, such as example code, can be included by indenting it with a tab.

Example:

```go
/*
MyFunction is an example function. It demonstrates how to use
documentation comments in Go.

Here is an example of how to use it:

    result := MyFunction()
    fmt.Println(result)
*/
func MyFunction() {
	// Function implementation
}
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)

## Documenting Packages

To document a package, place a documentation comment directly above the package clause. This should provide a high-level overview of what the package does and how to use it. It's also a good place to provide any necessary context or background information.

Example:

```go
/*
Package mypackage provides some useful utilities.

This package includes functions for string manipulation, sorting, and more.
For more details, see the documentation for each function.
*/
package mypackage
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)

## Documenting Functions

To document a function, place a documentation comment directly above the function declaration. This should describe what the function does, its parameters, its return values, and any errors it might return. It should also mention any side effects or conditions that callers should be aware of.

Example:

```go
/*
MyFunction performs some complex calculations.

It takes two parameters: an integer and a string. It returns the result
as a float64 and an error. If the string cannot be parsed as a float64,
the error will be non-nil.

The function does not modify its arguments or any global state.
*/
func MyFunction(a int, b string) (float64, error) {
	// Function implementation
}
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)

## Documenting Constants, Variables, and Types

To document a constant, variable, or type, place a documentation comment directly above its declaration. This should describe what it is and how it should be used. For variables that can be modified, document any constraints on their values or any effects that modifying them might have.

Example:

```go
/*
MyConstant is used for demonstration purposes.

This constant is used in various examples throughout this package.
Its value is "Hello, world!" and it should not be modified.
*/
const MyConstant = "Hello, world!"
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)

## Adding Examples

Go has a convention for adding examples to your documentation comments. If a function, type, method, or package has an example, it should be in a function with a name starting with "Example", and it should be in the form of a valid Go function that can be executed.

Example:

```go
/*
MyFunction performs some complex calculations.



It takes two parameters: an integer and a string. It returns the result
as a float64 and an error. If the string cannot be parsed as a float64,
the error will be non-nil.

Here is an example of its usage:

    result, err := MyFunction(5, "2.5")
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(result)
    }
*/
func MyFunction(a int, b string) (float64, error) {
	// Function implementation
}

// ExampleMyFunction demonstrates how to use MyFunction.
func ExampleMyFunction() {
	result, err := MyFunction(5, "2.5")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)

## Using go doc and godoc

The `go doc` and `godoc` command-line tools can be used to view the documentation for your code. They display the documentation comments in a readable format and can help ensure that your comments are well-formed and complete.

To view the documentation for a package:

```shell
go doc
```

To view the documentation for a specific item:

```shell
go doc MyFunction
```

`godoc` can also be used to serve a local web server that displays your documentation in a web browser:

```shell
godoc -http=:6060
```
[(back-to-top)](#go-documentation-commentaries-cheat-sheet)