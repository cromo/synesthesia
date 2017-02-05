package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const version = "0.1.0+Go"

type ansiColorer func(color int) string

type options struct {
	foreground ansiColorer
	background ansiColorer
	pattern    string
}

type colorMode int

const (
	mode16Color colorMode = iota
	mode256color
)

func main() {
	interactLinewise(colorLineMaker(parseArgs(os.Args[1:])))
}

func parseArgs(args []string) options {
	colorForeground, colorBackground := true, false
	regexps := make([]string, 0)
	colorDepth := mode256color

	acceptingFlags := true
	for _, arg := range args {
		if !acceptingFlags {
			regexps = append(regexps, arg)
			continue
		}

		if arg == "-h" || arg == "--help" {
			printUsage()
			os.Exit(0)
		} else if arg == "-v" || arg == "--version" {
			fmt.Printf("Synesthesia version %v\n", version)
			os.Exit(0)
		} else if arg == "-f" || arg == "--foreground" {
			colorForeground = true
		} else if arg == "--no-foreground" {
			colorForeground = false
		} else if arg == "-b" || arg == "--background" {
			colorBackground = true
		} else if arg == "--no-background" {
			colorBackground = false
		} else if arg == "--16" {
			colorDepth = mode16Color
		} else if arg == "--256" {
			colorDepth = mode256color
		} else if arg == "--" {
			acceptingFlags = false
		} else if strings.HasPrefix(arg, "-") {
			fmt.Printf("Unrecognized flag: %s\n", arg)
			printUsage()
			os.Exit(1)
		} else {
			regexps = append(regexps, arg)
		}
	}

	foregroundColorer, backgroundColorer := colorNoop, colorNoop
	if colorDepth == mode16Color {
		if colorForeground {
			foregroundColorer = colorAnsi16Foreground
		}
		if colorBackground {
			backgroundColorer = colorAnsi16Background
		}
	} else if colorDepth == mode256color {
		if colorForeground {
			foregroundColorer = colorAnsi256Foreground
		}
		if colorBackground {
			backgroundColorer = colorAnsi256Background
		}
	}

	validateRegexps(regexps)
	for i, pattern := range regexps {
		regexps[i] = fmt.Sprintf("(?:%v)", pattern)
	}

	return options{foregroundColorer, backgroundColorer, strings.Join(regexps, "|")}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %v ([flags]|[patterns])... [-- [patterns]...]
Color standard input based on the values matched by the provided regular
expressions.
Before --, any parameters starting with a dash will be interpreted as flags.
After it, all parameters are interpreted as patterns.
A pattern is any valid Go regular expression.
Multiples of flags controlling the same option are allowed; the last flag
specified will take precedence over the ones that come before it.
Flags:
-b --background     Turn on or off background colorization. (default off)
   --no-background
-f --foreground     Turn on or off foreground colorization. (default on)
   --no-foreground
-h --help           Print usage information.
`, os.Args[0])
}

func validateRegexps(regexps []string) {
	valid := true
	for _, pattern := range regexps {
		if _, err := regexp.Compile(pattern); err != nil {
			valid = false
			fmt.Fprintf(os.Stderr, "Invalid regex: %v\n", pattern)
		}
	}
	if !valid {
		os.Exit(1)
	}
}

func interactLinewise(transform func([]byte) []byte) {
	input := bufio.NewReader(os.Stdin)
	line, err := input.ReadBytes('\n')
	for err == nil {
		fmt.Print(string(transform(line)))
		line, err = input.ReadBytes('\n')
	}
	fmt.Print(string(transform(line)))
}

func colorLineMaker(options options) func([]byte) []byte {
	pattern := regexp.MustCompile(options.pattern)
	return func(line []byte) []byte {
		return colorPatternInString(pattern, line, options.foreground, options.background)
	}
}

func colorPatternInString(pattern *regexp.Regexp, buf []byte, foreground, background ansiColorer) []byte {
	return pattern.ReplaceAllFunc(buf, colorMatchMaker(foreground, background))
}

func colorMatchMaker(foreground, background ansiColorer) func([]byte) []byte {
	return func(match []byte) []byte {
		return color(match, foreground, background)
	}
}

func color(buf []byte, foreground, background ansiColorer) []byte {
	sum := md5.Sum(buf)
	color := int(sum[md5.Size-3])<<16 | int(sum[md5.Size-2])<<8 | int(sum[md5.Size-1])
	return []byte(fmt.Sprintf("%v%v%v\033[m", foreground(color), background(0xFFFFFF-color), string(buf)))
}

func colorNoop(color int) string {
	return ""
}

func colorAnsi16Foreground(color int) string {
	color = color % 16
	index := color%8 + 30
	if color < 8 {
		return fmt.Sprintf("\033[%vm", index)
	}
	return fmt.Sprintf("\033[%v;1m", index)
}

func colorAnsi16Background(color int) string {
	color = color % 16
	index := color%8 + 40
	if color < 8 {
		return fmt.Sprintf("\033[%vm", index)
	}
	return fmt.Sprintf("\033[%v;1m", index)
}

func colorAnsi256Foreground(color int) string {
	r, g, b := color>>(16+3), color>>(8+3)&0x1F, color>>3&0x1F
	return fmt.Sprintf("\033[38;5;%vm", rgbTo216Color(r, g, b)+16)
}

func colorAnsi256Background(color int) string {
	r, g, b := color>>(16+3), color>>(8+3)&0x1F, color>>3&0x1F
	return fmt.Sprintf("\033[48;5;%vm", rgbTo216Color(r, g, b)+16)
}

func rgbFrom216Color(color int) (int, int, int) {
	return color / 36, color / 6 % 6, color % 6
}

func rgbTo216Color(r, g, b int) int {
	return r*36 + g*6 + b
}
