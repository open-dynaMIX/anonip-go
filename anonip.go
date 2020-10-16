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
		maskedIp := maskIP(ip, args)
		if args.Increment > 0 {
			incrementIP(maskedIp, args.Increment)
		}
		line = strings.ReplaceAll(line, ipString, maskedIp.String())
	}
	channel <- line
}

type Args struct {
	IpV4Mask  int     `arg:"-4,--ipv4mask" default:"12" placeholder:"INTEGER" help:"truncate the last n bits"`
	IpV6Mask  int     `arg:"-6,--ipv6mask" default:"84" placeholder:"INTEGER" help:"truncate the last n bits"`
	Increment uint    `arg:"-i,--increment" default:"0" placeholder:"INTEGER" help:"increment the IP address by n"`
	Columns   []uint  `arg:"-c,--columns" placeholder:"INTEGER [INTEGER ...]" help:"assume IP address is in column n (1-based indexed) [default: 0]"`
	Delimiter string  `arg:"-l,--delimiter" default:" " placeholder:"STRING" help:"log delimiter"`
	Replace   *string `arg:"-r,--replace" placeholder:"STRING" help:"replacement string in case address parsing fails (Example: 0.0.0.0)"`
}

func parseArgs() (Args, *arg.Parser, error) {
	var args Args
	p := arg.MustParse(&args)
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
		osExit(-1)
	}
	run(args)
}
