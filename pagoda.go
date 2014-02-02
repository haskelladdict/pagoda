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
//  "reflect"
  "strconv"
  "unicode/utf8"
//  "log"
//  "io/ioutil"
)


// parse_spec is the parent type describing all options and usage
type parseSpec struct {
  Usage_info string
  Options []jsonOption    // options according to the spec
}


// const used for type tagging
const (
  aBool = iota
  anInt
  aFloat
  aString
)


// json_option describes a single option specification as parsed from the
// JSON description
type jsonOption struct {
  Short_option string
  Long_option string
  Description string
  Type string
  Default *string      // use a pointer to be able to distinguish the empty 
                       // string from non-present option
  value interface{}
}


// Usage prints the usage information for the package
func (p *parseSpec) Usage() {
  fmt.Printf("Usage: %s %s\n", os.Args[0], p.Usage_info)
  fmt.Println()
  for _, opt := range p.Options {
    fmt.Printf("\t-%s  --%s  %s", opt.Short_option, opt.Long_option,
      opt.Description)
    if opt.Default != nil {
      fmt.Printf("  [default: %s]", *opt.Default)
    }
    fmt.Printf("\n")
  }
}


// Init parses the specification of option in JSON format
func Init(content []byte) (*parseSpec, error) {

  var parse_info parseSpec
  err := json.Unmarshal(content, &parse_info)
  if err != nil {
    return nil, err
  }

  err = validate_specs(&parse_info)
  if err != nil {
    return nil, err
  }

  err = extract_defaults(&parse_info)
  if err != nil {
    return nil, err
  }

  err = match_spec_to_args(&parse_info, os.Args)
  if err != nil {
    return nil, err
  }

  fmt.Println(parse_info)
  return &parse_info, nil
}


// validate_defaults checks that a usage string was given and 
// that each spec has at least a short or a long option
func validate_specs(parse_info *parseSpec) error {
  if parse_info.Usage_info == "" {
    return fmt.Errorf("Usage string missing")
  }

  for _, opt := range parse_info.Options {
    if opt.Short_option == "" && opt.Long_option == "" {
      return fmt.Errorf("Need at least a short or long description.")
    }

    if opt.Type == "" {
      return fmt.Errorf("Need a type descriptor for option %s.",
        opt.Short_option)
    }
  }

  return nil
}


// extract_defaults looks at the default field (if present) and
// attempts to determine the type of the option field
func extract_defaults(parse_info *parseSpec) error {

  opts := parse_info.Options
  for i := 0; i < len(opts); i++ {
    if opts[i].Default != nil {
      val, err := string_to_type(*opts[i].Default, opts[i].Type)
      if err != nil {
        return err
      }

      opts[i].value = val
    }
  }

  return nil
}


// string_to_type converts value to the requested type and strows
// an error if that fails
func string_to_type(value string, theType string) (interface{}, error) {

  switch theType {
  case "bool":
    if value == "true" {
      return true, nil
    } else if value == "false" {
      return false, nil
    } else {
      return nil, fmt.Errorf("cannot convert %s to requested type %s",
        value, theType)
    }

  case "int":
    if i, err := strconv.Atoi(value); err == nil {
      return i, nil
    } else {
      return nil, fmt.Errorf("cannot convert %s to requested type %s",
        value, theType)
    }

  case "float":
    if v, err := strconv.ParseFloat(value, 64); err == nil {
      return v, nil
    } else {
      return nil, fmt.Errorf("cannot convert %s to requested type %s", 
        value, theType)
    }

  case "string":
    return value, nil

  default:
    return nil, fmt.Errorf("unknow type %s", theType)
  }
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
func find_parse_spec(spec *parseSpec, name string) (*jsonOption, bool) {

  opts := spec.Options
  for i := 0; i < len(opts); i++ {
    if opts[i].Short_option == name || opts[i].Long_option == name {
      return &opts[i], true
    }
  }

  return &jsonOption{}, false
}


// match_spec_to_args matches a parse_info spec to the provided command
// line options. Entries in parse_info which are lacking are ignored.
// If the command line contains entries which are not in the spec the
// function throws an error.
func match_spec_to_args(parsed *parseSpec, args []string) error {

  var opt_name, opt_val string
  var ok bool
  for i := 1; i < len(args); i++ {
    opt_name, opt_val, ok = decode_option(args[i])
    if !ok {
      return fmt.Errorf("Encountered unknow commandline toke %s", args[i])
    }

    opt_spec, ok := find_parse_spec(parsed, opt_name)
    if !ok {
      continue
    }

    // if the option is not of type bool and we don't have
    // a value yet we peak at the next arg if present
    if opt_val == "" && opt_spec.Type != "bool" {
      if i+1 < len(args) {
        if _, _, ok := decode_option(args[i+1]); !ok {
          i++
          opt_val = args[i]
        }
      }
    }

    // check that we got a value if the option doesn't have default 
    if opt_spec.Default == nil && opt_val == "" {
      return fmt.Errorf("Missing value for option %s", opt_spec.Short_option)
    }

    // check that the provided option has the correct typ
    if opt_val != "" {
      val, err := string_to_type(opt_val, opt_spec.Type)
      if err != nil {
        return err
      }
      opt_spec.value = val
    }

    //fmt.Println("option ****", opt_name, opt_val, opt_spec.value)
  }
  return nil
}



