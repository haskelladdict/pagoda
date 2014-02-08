// Copyright 2014 Markus Dittrich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// this is a simple example for the pagoda commandline
// parsing library
package main

import (
  "fmt"
  "log"
  "github.com/haskelladdict/pagoda"
)


var specs = []byte(`
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

  "Usage_info" : "[options] <filename>"
}
`)




func main() {

  var err error
  flags, err := pagoda.Init(specs)
  if err != nil {
    log.Fatal(err)
  }

  aVal, err := flags.Value("a")
  if err != nil {
    //flags.Usage()
    return
  }
  fmt.Printf("option -a is set to %v\n", aVal)

  var bVal interface{}
  bVal, err = flags.Value("b")
  if err != nil {
    fmt.Println(err)
    //flags.Usage()
    return
  }
  fmt.Printf("received value %f for b\n", bVal.(float64))

  fmt.Printf("the remaining command line options are %v\n", flags.Remainder())
}
