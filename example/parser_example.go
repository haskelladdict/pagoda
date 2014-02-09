// Copyright 2014 Markus Dittrich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// this is a simple example for the pagoda commandline
// parsing library
package main

import (
  "fmt"
//  "log"
  "os"
  "github.com/haskelladdict/pagoda"
)


var specs1 = []byte(`
{
  "options" : [
  {
    "Short_option" : "a",
    "Long_option"  : "all",
    "Description"  : "list them all",
    "Type"         : "bool",
    "Default"      : "true"
  },
  { 
    "Short_option" : "b",
    "Long_option"  : "bar",
    "Description"  : "this causes trouble",
    "Type"         : "float"
  },
  { 
    "Short_option" : "c",
    "Description" : "the name of the thing",
    "Type"        : "string"
  }

  ],

  "Usage_info" : "parser_example is a program for testing pagoda"
}
`)


var specs = []byte(`
{
  "options" : [
  {
    "short_option" : "a",
    "long_option"  : "all",
    "description"  : "list them all",
    "type"         : "bool",
    "default"      : "true",
    "subcommand"   : "general"
  },
  { 
    "short_option" : "b",
    "long_option"  : "bloat",
    "description"  : "this causes trouble",
    "type"         : "int",
    "subcommand"    : "general"
  },
  { 
    "short_option" : "c",
    "description"  : "the name of the thing",
    "type"         : "string",
    "subcommand"    : "special"
  }

  ],

  "usage_info" : "parser_example is a program for testing pagoda",

  "subcommand_info" : [
    { "general" : "this is some general description" },
    { "special" : "this is really special" }
  ]
}
`)




func main() {

  var err error
  flags, err := pagoda.Init(specs)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  if len(os.Args) <= 1 {
    fmt.Println(flags.Usage())
    os.Exit(1)
  }


  if f, _:= flags.Subcommand(); f == "general" {
    aVal, err := flags.Value("a")
    if err != nil {
      fmt.Println("a was not provided")
      //fmt.Println(err)
      //os.Exit(1)
    }
    fmt.Printf("option -a is set to %v\n", aVal)

    //var bVal interface{}
    _ , err = flags.Value("b")
    if err != nil {
      //fmt.Println(err)
      fmt.Println("b was not provided")
      os.Exit(1)
    }
  } else if f, _ := flags.Subcommand(); f == "special" {
    _ , err := flags.Value("c")
    if err != nil {
      fmt.Println("no c option was provided")
    }
  }

  fmt.Printf("the remaining command line options are %v\n", flags.Remainder())
}
