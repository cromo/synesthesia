package main

import "bufio"
import "fmt"
import "crypto/md5"
import "os"
import "regexp"
import "strings"

const version = "0.1.0+Go"

type options struct {
	foreground bool
	background bool
	pattern    string
}

func main() {
	interactLinewise(colorLineMaker(parseArgs(os.Args[1:])))
}

func parseArgs(args []string) options {
	options := options{true, false, ""}
	regexps := make([]string, 0)
	acceptingFlags := true
	for _, arg := range args {
		if acceptingFlags {
			if arg == "-h" || arg == "--help" {
				printUsage()
				os.Exit(0)
			} else if arg == "-v" || arg == "--version" {
				fmt.Printf("Synesthesia version %v\n", version)
				os.Exit(0)
			} else if arg == "-f" || arg == "--foreground" {
				options.foreground = true
			} else if arg == "--no-foreground" {
				options.foreground = false
			} else if arg == "-b" || arg == "--background" {
				options.background = true
			} else if arg == "--no-background" {
				options.background = false
			} else if arg == "--" {
				acceptingFlags = false
			} else if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unrecognized flag: %s\n", arg)
				printUsage()
				os.Exit(1)
			} else {
				regexps = append(regexps, arg)
			}
		} else {
			regexps = append(regexps, arg)
		}
	}

	validateRegexps(regexps)
	for i, pattern := range regexps {
		regexps[i] = fmt.Sprintf("(?:%v)", pattern)
	}
	options.pattern = strings.Join(regexps, "|")
	return options
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

func colorPatternInString(pattern *regexp.Regexp, buf []byte, foreground, background bool) []byte {
	return pattern.ReplaceAllFunc(buf, colorMatchMaker(foreground, background))
}

func colorMatchMaker(foreground, background bool) func([]byte) []byte {
	return func(match []byte) []byte {
		return color(match, foreground, background)
	}
}

func color(buf []byte, foreground, background bool) []byte {
	sum := md5.Sum(buf)
	r, g, b := rgbFrom216Color(int(sum[md5.Size-1] % 216))
	br, bg, bb := 5-r, 5-g, 5-b
	fore_color, back_color := "", ""
	if foreground {
		fore_color = fmt.Sprintf("\033[38;5;%vm", rgbTo216Color(r, g, b)+16)
	}
	if background {
		back_color = fmt.Sprintf("\033[48;5;%vm", rgbTo216Color(br, bg, bb)+16)
	}
	return []byte(fmt.Sprintf("%v%v%v\033[m", fore_color, back_color, string(buf)))
}

func rgbFrom216Color(color int) (int, int, int) {
	return color / 36, color / 6 % 6, color % 6
}

func rgbTo216Color(r, g, b int) int {
	return r*36 + g*6 + b
}
