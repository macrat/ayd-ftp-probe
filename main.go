package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/macrat/ayd/lib-ayd"
)

var (
	version     = "HEAD"
	commit      = "UNKNOWN"
	showVersion = flag.Bool("v", false, "show version info")
)

func init() {
	flag.Usage = func() {
		fmt.Println("FTP and FTPS protocol plugin for Ayd? status monitoring service.")
		fmt.Println()
		fmt.Printf("usage: %s TARGET_URL\n", filepath.Base(os.Args[0]))
	}
}

func NormalizeURL(u *url.URL) *url.URL {
	normalized := &url.URL{
		Scheme: u.Scheme,
		User:   u.User,
		Host:   u.Host,
	}
	if u.Opaque != "" {
		normalized.Host = u.Opaque
	}
	return normalized
}

func Check(logger ayd.Logger, target *url.URL) {
	options := []ftp.DialOption{
		ftp.DialWithTimeout(10 * time.Minute),
	}
	if target.Scheme == "ftps" {
		options = append(options, ftp.DialWithExplicitTLS(&tls.Config{}))
	}

	logger = logger.StartTimer()

	addr := target.Host
	if target.Port() == "" {
		addr += ":21"
	}
	conn, err := ftp.Dial(addr)
	if err != nil {
		logger.Failure(err.Error())
		return
	}
	defer conn.Quit()

	if target.User == nil {
		logger.Healthy("succeed connect")
		return
	}

	pass, _ := target.User.Password()
	err = conn.Login(target.User.Username(), pass)
	if err != nil {
		logger.Failure(err.Error())
		return
	}
	defer conn.Logout()

	logger.Healthy("succeed connect and login")
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ayd-ftp-probe %s (%s)\n", version, commit)
		return
	}

	args, err := ayd.ParseProbePluginArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "usage: %s TARGET_URL\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	target := NormalizeURL(args.TargetURL)
	logger := ayd.NewLogger(target)

	Check(logger, target)
}
