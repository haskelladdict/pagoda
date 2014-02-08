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
  "strconv"
  "unicode/utf8"
)


// parse_spec describes all options and usage actually present on the
// command line and also keeps track of any remaining command line entries
type parsedSpec struct {
  templateSpec
  cmdlOptions jsonOptions

  subcommand_mode bool
  subcommand string
  subcommand_options map[string]jsonOptions

  remainder []string      // unparsed remainder of command line
}


// target_spec describes all possible options and usage per json spec file
type templateSpec struct {
  Usage_info string
  Options jsonOptions   // options according to the spec
}


// json_option describes a single option specification as parsed from the
// JSON description
type jsonOption struct {
  Short_option string
  Long_option string
  Description string
  Type string
  Default *string      // use a pointer to be able to distinguish the empty 
                       // string from non-present option
  Subcommand *string
  value interface{}    // option value of type Type
}

type jsonOptions []jsonOption


// Value returns the value known for the given option. Either the
// long or short option can be provided.
// NOTE: The look up could be made more efficient via a map 
func (p *parsedSpec) Value(key string) (interface{}, error) {

  for _, opt := range p.cmdlOptions {
    if key == opt.Short_option || key == opt.Long_option {
      return opt.value, nil
    }
  }

  return nil, fmt.Errorf("Pagoda: command line option %s not found", key)
}


// Remainder returns a slice with all command line options that
// were not parsed into options
func (p *parsedSpec) Remainder() []string {
  return p.remainder
}


// Init parses the specification of option in JSON format
func Init(content []byte) (*parsedSpec, error) {

  var parse_info templateSpec
  err := json.Unmarshal(content, &parse_info)
  if err != nil {
    return nil, err
  }

  haveGroupMode := check_for_group_mode(&parse_info)

  err = validate_specs(&parse_info, haveGroupMode)
  if err != nil {
    return nil, err
  }

  err = extract_defaults(&parse_info)
  if err != nil {
    return nil, err
  }

  matched_info, err := match_spec_to_args(&parse_info, os.Args, haveGroupMode)
  if err != nil {
    return nil, err
  }

  // check if help was requested. In that case show Usage() and then exit
  if _, err := matched_info.Value("h"); err == nil {
    if matched_info.subcommand_mode {
      command_usage(matched_info.Usage_info, matched_info.subcommand,
        matched_info.Options, os.Args)
    } else {
      matched_info.Usage()
    }
    os.Exit(0)
  }

  return matched_info, nil
}


// Usage prints the usage information for the package
func (p *parsedSpec) Usage() {

  fmt.Println(p.Usage_info)
  fmt.Println()

  if p.subcommand_mode {
    fmt.Printf("Usage: %s SUBCOMMAND [arguments]\n", os.Args[0])
    fmt.Println("\nAvailable SUBCOMMANDS are:\n")
    for k, _ := range p.subcommand_options {
      fmt.Printf("\t%-10s\n", k)
    }
  } else {
    command_usage(p.Usage_info, "", p.Options, os.Args)
  }

  fmt.Println()
}


// command_usage prints the usage for a specific subcommand
func command_usage(info string, subcommand string, options jsonOptions,
  args []string) {

  fmt.Printf("Usage: %s %s [arguments]\n\n", args[0], subcommand)
  for _, opt := range options {

    if opt.Long_option == "" {
      fmt.Printf("\t-%-15s  %s", opt.Short_option, opt.Description)
    } else if opt.Short_option == "" {
      fmt.Printf("\t    --%-10s  %s", opt.Long_option, opt.Description)
    } else {
      fmt.Printf("\t-%s  --%-10s  %s", opt.Short_option, opt.Long_option,
        opt.Description)
    }

    if opt.Default != nil && opt.Type != "bool" {
      fmt.Printf("  [default: %s]", *opt.Default)
    }
    fmt.Printf("\n")
  }
}


// check_for_group_mode tests if at least one option has a 
// group mode set
func check_for_group_mode(parse_info *templateSpec) bool {

  haveGroupMode := false
  for _, item := range parse_info.Options {
    if item.Subcommand != nil {
      haveGroupMode = true
      break
    }
  }

  return haveGroupMode
}


// validate_defaults checks that a usage string was given and 
// that each spec has at least a short or a long option. 
func validate_specs(parse_info *templateSpec, haveGroupMode bool) error {
  if parse_info.Usage_info == "" {
    return fmt.Errorf("Pagoda: Usage string missing")
  }

  for _, opt := range parse_info.Options {
    if opt.Short_option == "" && opt.Long_option == "" {
      return fmt.Errorf("Pagoda: Need at least a short or long description.")
    }

    if opt.Type == "" {
      return fmt.Errorf("Pagoda: Need a type descriptor for option %s.",
        opt.Short_option)
    }

    if haveGroupMode && opt.Subcommand == nil {
      return fmt.Errorf("Pagoda: Only some options have command group " +
        "mode selected.")
    }
  }

  return nil
}


// extract_defaults extracts the default field in the proper type
func extract_defaults(parse_info *templateSpec) error {

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
      return nil, fmt.Errorf("Pagoda: cannot convert %s to requested type %s",
        value, theType)
    }

  case "float":
    if v, err := strconv.ParseFloat(value, 64); err == nil {
      return v, nil
    } else {
      return nil, fmt.Errorf("Pagoda: cannot convert %s to requested type %s",
        value, theType)
    }

  case "string":
    return value, nil

  default:
    return nil, fmt.Errorf("Pagoda: unknow type %s", theType)
  }
}



// inject_default_help_option adds a default help option to the list of
// command line switches
func inject_default_help_option(opts *jsonOptions) {

  helpDefault := "true"
  *opts = append(*opts,
    jsonOption{"h", "help", "print this help text and exit", "bool",
      &helpDefault, nil, nil})
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
func find_parse_spec(opts jsonOptions, name string) (*jsonOption, bool) {

  for i := 0; i < len(opts); i++ {
    if opts[i].Short_option == name || opts[i].Long_option == name {
      return &opts[i], true
    }
  }

  return &jsonOption{}, false
}


// group_options distributes all options across a map with group
// name as key
func group_options(parse_data *parsedSpec, template *templateSpec) map[string]jsonOptions {

  subcommand_options := make(map[string]jsonOptions)
  for _, opts := range template.Options {
    key := *opts.Subcommand
    if _, ok := subcommand_options[key]; !ok {
      option_list := make(jsonOptions, 0)
      option_list = append(option_list, opts)
      subcommand_options[key] = option_list
    } else {
      subcommand_options[key] = append(subcommand_options[key], opts)
    }
  }
  return subcommand_options
}


// inialize_parsed_spec initialize the parsed spec
func initialize_parsed_spec(haveSubcommands bool, template *templateSpec) *parsedSpec {

  // initialize final parsed specs
  parsed := parsedSpec{}
  parsed.Options = template.Options
  parsed.cmdlOptions = make([]jsonOption, 0)
  parsed.Usage_info = template.Usage_info

  if haveSubcommands {
    parsed.subcommand_options = group_options(&parsed, template)
    parsed.subcommand_mode= haveSubcommands
  }

  return &parsed
}


func extract_subcommand(parsed *parsedSpec, args []string) ([]string, error) {

  subcommand := args[1]
  opts, ok := parsed.subcommand_options[subcommand];
  if !ok {
    return nil, fmt.Errorf("Pagoda: Unknown command group %s", subcommand)
  }

  parsed.Options = opts
  parsed.subcommand = subcommand
  return args[2:], nil
}


// match_spec_to_args matches a parse_info spec to the provided command
// line options. Entries in parse_info which are lacking are ignored.
// If the command line contains entries which are not in the spec the
// function throws an error.
func match_spec_to_args(template *templateSpec, args []string,
  haveSubcommands bool) (*parsedSpec, error) {

  parsed := initialize_parsed_spec(haveSubcommands, template)

  if len(args) <= 1 {
    return parsed, nil
  }

  if haveSubcommands {
    var err error
    args, err = extract_subcommand(parsed, args)
    if err != nil {
      parsed.Usage()
      return nil, err
    }
  } else {
    args = args[1:]
  }

  inject_default_help_option(&parsed.Options)

  var opt_name, opt_val string
  var ok bool
  for i := 0; i < len(args); i++ {
    opt_name, opt_val, ok = decode_option(args[i])
    if !ok {
      parsed.remainder = args[i:]
      break
    }

    opt_spec, ok := find_parse_spec(parsed.Options, opt_name)
    if !ok {
      parsed.Usage()
      return nil, fmt.Errorf("Pagoda: Unknown command line option %s",
        args[i])
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
      parsed.Usage()
      return nil, fmt.Errorf("Pagoda: Missing value for option %s",
        opt_spec.Short_option)
    }

    // check that the provided option has the correct type
    if opt_val != "" {
      val, err := string_to_type(opt_val, opt_spec.Type)
      if err != nil {
        return nil, err
      }
      opt_spec.value = val
    }

    parsed.cmdlOptions = append(parsed.cmdlOptions, *opt_spec)
  }
  return parsed, nil
}
