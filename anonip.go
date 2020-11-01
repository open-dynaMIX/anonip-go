package main

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)
import "bufio"
import "net"
import "github.com/alexflint/go-arg"

var version = "0.0.0-alpha.1"

// to enable monkey-patching during tests
var osExit = os.Exit
var defaultLogWriter = os.Stdout
var defaultLogReader io.Reader = os.Stdin

var privateIPBlocksStrings = []string{
	"127.0.0.0/8",    // IPv4 loopback
	"10.0.0.0/8",     // RFC1918
	"172.16.0.0/12",  // RFC1918
	"192.168.0.0/16", // RFC1918
	"169.254.0.0/16", // RFC3927 link-local
	"::1/128",        // IPv6 loopback
	"fe80::/10",      // IPv6 link-local
	"fc00::/7",       // IPv6 unique local addr
}

// OpenFile is a wrapper around os.OpenFile for better control in tests
func OpenFile(name string, flag int, perm os.FileMode) *os.File {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		logError(err)
		osExit(-1)
		return nil // just in case osExit was monkey-patched
	}
	return f
}

var privateIPBlocks []*net.IPNet

func initPrivateIPBlocks() {
	for _, cidr := range privateIPBlocksStrings {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			logError(err)
			osExit(2)
			return // just in case osExit was monkey-patched
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func maskIP(ip net.IP, args Args) net.IP {
	if ip := ip.To4(); ip != nil {
		mask := net.CIDRMask(32-args.IPV4Mask, 32)
		return ip.Mask(mask)
	}
	mask := net.CIDRMask(128-args.IPV6Mask, 128)
	return ip.Mask(mask)
}

func incrementIP(ip net.IP, amount uint) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i] += byte(amount)
		if ip[i] != 0 {
			break
		}
	}
}

func _trimBrackets(ipString string) (string, net.IP) {
	ipString = strings.Trim(ipString, "[]")
	return ipString, net.ParseIP(ipString)
}

func _handlePort(ipString string) (string, net.IP) {
	strippedIPString, _, err := net.SplitHostPort(ipString)
	if err != nil {
		parts := strings.Split(ipString, "]")
		if len(parts) > 1 {
			return parts[0], net.ParseIP(parts[0])
		}
		return ipString, nil
	}

	return strippedIPString, net.ParseIP(strippedIPString)
}

func getIP(ipString string) (string, net.IP) {
	ip := net.ParseIP(ipString)
	if ip == nil {
		ipString, ip = _trimBrackets(ipString)
		if ip == nil {
			return _handlePort(ipString)
		}
		return ipString, ip
	}
	return ipString, ip
}

func getIPStringsRegex(line string, regex *regexp.Regexp) []string {
	return regex.FindStringSubmatch(line)
}

func getIPStringsColumn(line string, columns []uint, delimiter string) []string {
	logList := strings.Split(line, delimiter)
	ipList := []string{}
	for _, column := range columns {
		if int(column) > len(logList)-1 {
			continue
		}
		ipList = append(ipList, logList[column])
	}
	return ipList
}

func printLog(w io.Writer, line string) {
	_, err := w.Write([]byte(line + "\n"))
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	_, _ = os.Stderr.WriteString("error: " + err.Error() + "\n")
}

func handleLine(line string, args Args, channel chan string) {
	if line == "" {
		channel <- line
		return
	}
	var ipStrings []string
	if args.Regex != nil {
		ipStrings = getIPStringsRegex(line, args.Regex)
	} else {
		ipStrings = getIPStringsColumn(line, args.Columns, args.Delimiter)
	}
	for _, ipString := range ipStrings {
		ipString, ip := getIP(ipString)
		if ip == nil {
			if args.Replace != nil {
				line = strings.Replace(line, ipString, *args.Replace, 1)
			}
			continue
		}
		if args.SkipPrivate {
			if isPrivateIP(ip) {
				continue
			}
		}
		maskedIP := maskIP(ip, args)
		if args.Increment > 0 {
			incrementIP(maskedIP, args.Increment)
		}
		line = strings.ReplaceAll(line, ipString, maskedIP.String())
	}
	channel <- line
}

// Args will hold parsed CLI arguments
type Args struct {
	IPV4Mask    int            `arg:"-4,--ipv4mask,env" default:"12" placeholder:"INTEGER" help:"truncate the last n bits"`
	IPV6Mask    int            `arg:"-6,--ipv6mask,env" default:"84" placeholder:"INTEGER" help:"truncate the last n bits"`
	Increment   uint           `arg:"-i,--increment,env" default:"0" placeholder:"INTEGER" help:"increment the IP address by n"`
	RawOutput   string         `arg:"-o,--output,env" placeholder:"FILE" help:"file or FIFO to write to [default: stdout]"`
	Output      io.Writer      `arg:"-"`
	RawInput    string         `arg:"--input,env" placeholder:"FILE" help:"file or FIFO to read from [default: stdin]"`
	Input       io.Reader      `arg:"-"`
	Columns     []uint         `arg:"-c,--columns,env" placeholder:"INTEGER [INTEGER ...]" help:"assume IP address is in column n (1-based indexed) [default: 0]"`
	Delimiter   string         `arg:"-l,--delimiter,env" default:" " placeholder:"STRING" help:"log delimiter"`
	Replace     *string        `arg:"-r,--replace,env" placeholder:"STRING" help:"replacement string in case address parsing fails (Example: 0.0.0.0)"`
	RawRegex    []string       `arg:"--regex,env" placeholder:"STRING [STRING ...]" help:"regex"`
	Regex       *regexp.Regexp `arg:"-"`
	SkipPrivate bool           `arg:"-p,--skip-private,env" default:"false" help:"do not mask addresses in private ranges. See IANA Special-Purpose Address Registry"`
	Version     bool           `arg:"-v,--version" default:"false" help:"show program's version number and exit"`
}

func parseArgs() (Args, *arg.Parser, error) {
	var args Args
	p := arg.MustParse(&args)

	if args.Version {
		printLog(defaultLogWriter, version)
		osExit(0)
	}

	args.Output = defaultLogWriter
	if output := strings.Trim(args.RawOutput, " "); output != "" {
		file := OpenFile(args.RawOutput, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		args.Output = file
	}

	args.Input = defaultLogReader
	if input := strings.Trim(args.RawInput, " "); input != "" {
		file := OpenFile(args.RawInput, os.O_RDONLY, 0)
		args.Input = file
	}

	if args.IPV4Mask < 1 || args.IPV4Mask > 32 {
		return args, p, errors.New("argument -4/--ipv4mask: must be an integer between 1 and 32")
	}
	if args.IPV6Mask < 1 || args.IPV6Mask > 128 {
		return args, p, errors.New("argument -6/--ipv6mask: must be an integer between 1 and 128")
	}

	if len(args.RawRegex) != 0 {
		r, err := regexp.Compile(strings.Join(args.RawRegex, "|"))
		if err != nil {
			return args, p, errors.New("argument --regex: must be a valid regex string")
		}
		args.Regex = r
	}
	if len(args.Columns) == 0 {
		args.Columns = append(args.Columns, 0)
	} else {
		for i, col := range args.Columns {
			if col == 0 {
				return args, p, errors.New("column is 1-based indexed and must be > 0")
			}
			args.Columns[i]--
		}
	}
	return args, p, nil
}

func run(args Args) {
	if args.SkipPrivate {
		initPrivateIPBlocks()
	}
	channel := make(chan string)
	scanner := bufio.NewScanner(args.Input)
	for scanner.Scan() {
		go handleLine(scanner.Text(), args, channel)
		printLog(args.Output, <-channel)
	}
	if err := scanner.Err(); err != nil {
		logError(err)
		osExit(-1)
		return // just in case osExit was monkey-patched
	}
}

func main() {
	args, p, err := parseArgs()
	if err != nil {
		p.WriteUsage(os.Stderr)
		logError(err)
		osExit(-1)
		return // just in case osExit was monkey-patched
	}
	run(args)
}
