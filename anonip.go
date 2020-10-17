package main

import (
	"errors"
	"io"
	"os"
	"strings"
)
import "bufio"
import "log"
import "net"
import "github.com/alexflint/go-arg"

// to enable monkey-patching during tests
var osExit = os.Exit
var logWriter = os.Stdout
var logReader io.Reader = os.Stdin
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

// Wrapper around os.OpenFile for better control in tests
func OpenFile(name string, flag int, perm os.FileMode) *os.File {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		log.Println("error:", err)
		osExit(1)
	}
	return f
}

var privateIPBlocks []*net.IPNet

func initPrivateIPBlocks() {
	for _, cidr := range privateIPBlocksStrings {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Println("error:", err)
			osExit(2)
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
		mask := net.CIDRMask(32-args.IpV4Mask, 32)
		return ip.Mask(mask)
	}
	mask := net.CIDRMask(128-args.IpV6Mask, 128)
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
	strippedIpString, _, err := net.SplitHostPort(ipString)
	if err != nil {
		parts := strings.Split(ipString, "]")
		if len(parts) > 1 {
			return parts[0], net.ParseIP(parts[0])
		}
		return ipString, nil
	}

	return strippedIpString, net.ParseIP(strippedIpString)
}

func getIP(ipString string) (string, net.IP) {
	ip := net.ParseIP(ipString)
	if ip == nil {
		ipString, ip = _trimBrackets(ipString)
		ip := net.ParseIP(ipString)
		if ip == nil {
			return _handlePort(ipString)
		}
		return ipString, ip
	}
	return ipString, ip
}

func getIPStrings(line string, columns []uint, delimiter string) []string {
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
	w.Write([]byte(line + "\n"))
}

func handleLine(line string, args Args, channel chan string) {
	if line == "" {
		channel <- line
		return
	}
	ipStrings := getIPStrings(line, args.Columns, args.Delimiter)
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
		maskedIp := maskIP(ip, args)
		if args.Increment > 0 {
			incrementIP(maskedIp, args.Increment)
		}
		line = strings.ReplaceAll(line, ipString, maskedIp.String())
	}
	channel <- line
}

type Args struct {
	IpV4Mask    int     `arg:"-4,--ipv4mask" default:"12" placeholder:"INTEGER" help:"truncate the last n bits"`
	IpV6Mask    int     `arg:"-6,--ipv6mask" default:"84" placeholder:"INTEGER" help:"truncate the last n bits"`
	Increment   uint    `arg:"-i,--increment" default:"0" placeholder:"INTEGER" help:"increment the IP address by n"`
	Output      string  `arg:"-o,--output" placeholder:"FILE" help:"file or FIFO to write to [default: stdout]"`
	Input       string  `arg:"--input" placeholder:"FILE" help:"file or FIFO to read from [default: stdin]"`
	Columns     []uint  `arg:"-c,--columns" placeholder:"INTEGER [INTEGER ...]" help:"assume IP address is in column n (1-based indexed) [default: 0]"`
	Delimiter   string  `arg:"-l,--delimiter" default:" " placeholder:"STRING" help:"log delimiter"`
	Replace     *string `arg:"-r,--replace" placeholder:"STRING" help:"replacement string in case address parsing fails (Example: 0.0.0.0)"`
	SkipPrivate bool    `arg:"-p,--skip-private" default:"false" help:"do not mask addresses in private ranges. See IANA Special-Purpose Address Registry"`
}

func parseArgs() (Args, *arg.Parser, error) {
	var args Args
	p := arg.MustParse(&args)

	args.Output = strings.Trim(args.Output, " ")
	if args.Output != "" {
		file := OpenFile(args.Output, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		logWriter = file
	}

	args.Input = strings.Trim(args.Input, " ")
	if args.Input != "" {
		file := OpenFile(args.Input, os.O_RDONLY, 0)
		logReader = file
	}

	if args.IpV4Mask < 1 || args.IpV4Mask > 32 {
		return args, p, errors.New("argument -4/--ipv4mask: must be an integer between 1 and 32!")
	}
	if args.IpV6Mask < 1 || args.IpV6Mask > 128 {
		return args, p, errors.New("argument -6/--ipv6mask: must be an integer between 1 and 128!")
	}
	if len(args.Columns) == 0 {
		args.Columns = append(args.Columns, 0)
	} else {
		for i, col := range args.Columns {
			if col == 0 {
				return args, p, errors.New("Column is 1-based indexed and must be > 0!")
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
	scanner := bufio.NewScanner(logReader)
	for scanner.Scan() {
		go handleLine(scanner.Text(), args, channel)
		printLog(logWriter, <-channel)
	}
	if err := scanner.Err(); err != nil {
		log.Println("error:", err)
		osExit(1)
	}
}

func main() {
	args, p, err := parseArgs()
	if err != nil {
		p.WriteUsage(os.Stderr)
		log.Println("error:", err)
		osExit(1)
	}
	run(args)
}
