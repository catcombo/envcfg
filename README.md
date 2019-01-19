# EnvCfg

[![Build Status](https://travis-ci.org/catcombo/envcfg.svg)](https://travis-ci.org/catcombo/envcfg)
[![Go Report Card](https://goreportcard.com/badge/github.com/catcombo/envcfg)](https://goreportcard.com/report/github.com/catcombo/envcfg)
[![GoDoc](https://godoc.org/github.com/catcombo/envcfg?status.svg)](https://godoc.org/github.com/catcombo/envcfg)

Package `envcfg` provides functions to load values to a structure fields from `.env` file and from OS environment variables.

## Usage

Declare a structure and use tag `env` to define associated environment variable names for desired fields.

	type Cfg struct {
	    Debug       bool   `env:"DEBUG"`
	    DatabaseURL string `env:"DATABASE_URL"`
	}

Create a new structure to provide default values.

    cfg := Cfg{
        Debug: true,
        DatabaseURL: "sqlite:///db.sqlite",
    }

Call `envcfg.Load()` to load values from environment variables.

    err := envcfg.Load(&cfg)

Keep in mind that the values are first loaded from the `.env` file (if it exists) and then
from the OS environment variables that can override the values loaded from the file.

The syntax of the `.env` file should follow these rules:

 - Each line should be in VAR=VAL format
 - Lines beginning with # are processed as comments and ignored
 - Blank lines are ignored

## Notes

A limited number of field types are supported, but they should be enough for most cases.
Nested structures are supported, just mark nested fields with `env` tag as usual
(no special syntax for `.env` file required). To load .env file from a different location or
with a different name use `envcfg.LoadFile()`.
