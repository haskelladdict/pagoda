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
  "unicode/utf8"
  "log"
//  "io/ioutil"
)


// packet_options holds the parsed values of a command line
// spec at runtime
var parse_info parse_spec


// parse_spec is the parent type describing all options and usage
type parse_spec struct {
  Usage string
  Options []json_option    // options parsed from spec file
  ArgOptions []option      // options present and parsed from command line
}



// json_option describes a single option specification as parsed from the
// JSON description
type json_option struct {
  Short_option string
  Long_option string
  Description string
  Default string
}



// option describes an option based on json_option with value 
type option struct {
  json_option
  value interface{}
}



// Usage prints the usage information for the package
func Usage() {
  fmt.Printf("Usage: %s %s\n", os.Args[0], parse_info.Usage)
  fmt.Println()
  for _, opt := range parse_info.Options {
    fmt.Printf("\t-%s  --%s  %s\n", opt.Short_option, opt.Long_option,
      opt.Description)
  }
}



// parse_specs parses the specification of option in JSON format
func Parse_specs(content []byte) error {

  err := json.Unmarshal(content, &parse_info)
  if err != nil {
    return err
  }

  err = match_spec_to_args(parse_info, os.Args)
  if err != nil {
    return err
  }

  return nil
}



// test_for_option determines if a string is an option (i.e. starts
// either with a dash ('-') or a double dash ('--')). In that
// case it returns the name of the option and the value if the
// option was given via --opt=val.
func decode_option(item string) (string, string, error) {

  // check for dash
  c, s := utf8.DecodeRuneInString(item)
  if s == 0 || string(c) != "-" {
    return "", "", fmt.Errorf("%s is not an option\n", item)
  }
  i := s

  // skip next dash if present
  c, s = utf8.DecodeRuneInString(item[i:])
  if s != 0 && string(c) == "-" {
    i += s;
  }

  // scan until end or until we hit a "="
  opt := ""
  for i < len(item) {
    c, s = utf8.DecodeRuneInString(item[i:])
    i += s

    if s == 0 {
      return "", "", fmt.Errorf("failed to decode %s\n", item)
    } else if string(c) == "=" {
      break
    }

    opt += string(c)
  }

  val := ""
  for i < len(item) {
    c, s = utf8.DecodeRuneInString(item[i:])
    i += s

    if s == 0 {
      return "", "", fmt.Errorf("failed to decode %s\n", item)
    }

    val += string(c)
  }

  return opt, val, nil
}



// match_spec_to_args matches a parse_info spec to the provided command
// line options. Entries in parse_info which are lacking are ignored.
// If the command line contains entries which are not in the spec the
// function throws an error.
func match_spec_to_args(parsed parse_spec, args []string) error {
  fmt.Println("got it")

  i := 1
  for i < len(args) {

    currentArg := args[i]
    opt, val, err := decode_option(currentArg)
    if err != nil {
      log.Fatal(err)
    }

    fmt.Println("option ****", opt, val)

    i++
  }


  return nil
}



