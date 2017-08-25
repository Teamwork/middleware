[![Build Status](https://travis-ci.org/Teamwork/httperr.svg?branch=master)](https://travis-ci.org/Teamwork/httperr) [![Codecov](https://img.shields.io/codecov/c/github/Teamwork/httperr.svg?style=flat)](https://codecov.io/gh/Teamwork/httperr) [![GoDoc](https://godoc.org/github.com/Teamwork/httperr?status.svg)](http://godoc.org/github.com/Teamwork/httperr)

# httperr

The httperr package provides error-handling facilities for use with HTTP-centric
Go applications, such as HTTP servers and client libraries. It is heavily
influenced by, and intended for use along with the [github.com/pkg/errors](https://github.com/pkg/errors)
package.

The general use for the httperr package is to include an HTTP status code with
errors. This allows a clean, decoupling between logic buried deeply inside
your models or business logic, and the HTTP layer, which renders HTTP responses.

A new error message can be created with a status code as such:

    err := httperr.New(http.StatusNotFound, "file not found")

If you don't need a custom message, you may be content using one of the built-in
errors, which just returns a standard default error text.

    err := httperr.ErrNotFound

Or you may wish to decorate an existing error with an HTTP status code:

    if user, err := db.FetchUser(); err != nil {
        return httperr.WithStatus(http.StatusNotFound, err)
    }

This also means it is possible to override a previously set HTTP status code.
This can be useful if you want to return a NotFound when a database lookup fails,
but in some cases the same lookup should be treated as Forbidden:

    func Lookup() (*User,error) {
        // .. user not found
        return httperr.New(http.StatusNotFound, "user does not exist")
    }

    // .. somewhere else
    if err := Lookup() {
        return httperr.WithStatus(http.StatusForbidden, err)
    }

The error message will still be "user does not exist", but the HTTP status
code will be treated as 403 instead of 404.

To see what status code is tied to an error message, use the StatusCode method:

    status := httperr.StatusCode(err)

If err embeds an HTTP status code, it will be returned. If it does not, a 500
will be returned. This makes it reasonable to allow standard errors throughout
your application, but they will be treated as a 500/Internal Server Error if
they are not otherwise altered by the httperr package.
