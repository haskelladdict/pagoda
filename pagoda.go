// Copyright 2014 Markus Dittrich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// pagoda is a go library for command line parsing 
package pagoda

import (
  "encoding/json"
  "os"
  "fmt"
//  "log"
//  "io/ioutil"
)


// packet_options holds the parsed values of a command line
// spec at runtime
var parse_info parse_spec


// parse_spec is the parent type describing all options and usage
type parse_spec struct {
  Usage string
  Options []option
}



// option describes a single option specification
type option struct {
  Short_option string
  Long_option string
  Description string
}



// Usage prints the usage information for the package
func Usage() {
  fmt.Printf("Usage: %s %s\n", os.Args[0], parse_info.Usage)
  fmt.Println()
  for _, option := range parse_info.Options {
    fmt.Printf("\t%s  %s  %s\n", option.Short_option, option.Long_option,
      option.Description)
  }
}



// parse_specs parses the specification of option in JSON format
func Parse_specs(content []byte) error {

  err := json.Unmarshal(content, &parse_info)
  if err != nil {
    return err
  }

  return nil
}
