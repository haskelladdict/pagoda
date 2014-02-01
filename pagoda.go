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
//  "log"
//  "io/ioutil"
)


// parse_info holds the parsed values of a command line spec
var parse_info parseSpec


// option_info holds the filled in values of present on the actual
// command line value corresponding to a command line spec at runtime
var option_info optionSpec


// option_spec is the type describing a filled in view of all
// parsed commandline options which matched the parse_spec
type optionSpec struct {
  argOptions []option      // options present and parsed from command line
}


// parse_spec is the parent type describing all options and usage
type parseSpec struct {
  Usage string
  Options []jsonOption    // options according to the spec
}


// json_option describes a single option specification as parsed from the
// JSON description
type jsonOption struct {
  Short_option string
  Long_option string
  Description string
  Default string
}


// option describes an option based on json_option with value 
type option struct {
  jsonOption
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

  // initialize the option_spec
  option_info.argOptions = make([]option, 0)

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
func decode_option(item string) (string, string, bool) {

  // check for dash
  c, s := utf8.DecodeRuneInString(item)
  if s == 0 || string(c) != "-" {
    return "", "", false
  }
  i := s

  // skip next dash if present
  c, s = utf8.DecodeRuneInString(item[i:])
  if s != 0 && string(c) == "-" {
    i += s;
  }

  // scan until end of string or until we hit a "="
  opt := ""
  for i < len(item) {
    c, s = utf8.DecodeRuneInString(item[i:])
    i += s

    if s == 0 {
      return "", "", false
    } else if string(c) == "=" {
      break
    }

    opt += string(c)
  }

  // scan for optional value specified via opt=val
  val := ""
  for i < len(item) {
    c, s = utf8.DecodeRuneInString(item[i:])
    i += s

    if s == 0 {
      return "", "", false
    }

    val += string(c)
  }

  return opt, val, true
}



// find_option retrieves the parse_spec option entry corresponding 
// to the given name f present. Otherwise returns false.
func find_parse_spec(spec parseSpec, name string) (jsonOption, bool) {

  for _, opt := range spec.Options {
    if opt.Short_option == name || opt.Long_option == name {
      return opt, true
    }
  }

  return jsonOption{}, false
}


// match_spec_to_args matches a parse_info spec to the provided command
// line options. Entries in parse_info which are lacking are ignored.
// If the command line contains entries which are not in the spec the
// function throws an error.
func match_spec_to_args(parsed parseSpec, args []string) error {
  fmt.Println("got it")

  var opt_name, opt_val string
  var ok bool
  for i := 1; i < len(args); i++ {
    opt_name, opt_val, ok = decode_option(args[i])
    if !ok {
      continue
    }

    _, ok := find_parse_spec(parsed, opt_name)
    if !ok {
      continue
    }

    fmt.Println("option ****", opt_name, opt_val)
  }


  return nil
}



