package main_test

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"regexp"
	"testing"

	"github.com/macrat/ayd-ftp-probe"
	"github.com/macrat/ayd/lib-ayd"
	ftp "goftp.io/server/core"
)

type TestDriver struct{}

func (d TestDriver) Stat(s string) (ftp.FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (d TestDriver) ListDir(s string, f func(ftp.FileInfo) error) error {
	return nil
}

func (d TestDriver) DeleteDir(s string) error {
	return nil
}

func (d TestDriver) DeleteFile(s string) error {
	return nil
}

func (d TestDriver) Rename(x, y string) error {
	return nil
}

func (d TestDriver) MakeDir(s string) error {
	return nil
}

func (d TestDriver) GetFile(s string, i int64) (int64, io.ReadCloser, error) {
	return 0, nil, errors.New("not implemented")
}

func (d TestDriver) PutFile(s string, f io.Reader, b bool) (int64, error) {
	return 0, errors.New("not implemented")
}

func (d TestDriver) NewDriver() (ftp.Driver, error) {
	return d, nil
}

type TestAuth struct{}

func (a TestAuth) CheckPasswd(username, password string) (ok bool, err error) {
	if username == "hoge" && password == "fuga" {
		return true, nil
	}
	return false, nil
}

func StartTestServer(t *testing.T) *ftp.Server {
	t.Helper()
	server := ftp.NewServer(&ftp.ServerOpts{
		Factory: TestDriver{},
		Auth:    TestAuth{},
		Port:    21021,
		Logger:  &ftp.DiscardLogger{},
	})
	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Fatalf("failed to start ftp server: %s", err)
		}
		t.Cleanup(func() {
			server.Shutdown()
		})
	}()
	return server
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		Input, Output string
	}{
		{"ftp://localhost", "ftp://localhost"},
		{"ftp:localhost", "ftp://localhost"},
		{"ftps://foo@localhost", "ftps://foo@localhost"},
		{"ftp://foo:bar@localhost", "ftp://foo:bar@localhost"},
		{"ftp://foo:bar@localhost/path/to", "ftp://foo:bar@localhost"},
		{"ftps://foo:bar@localhost/path/to#fragment?query=abc", "ftps://foo:bar@localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.Input, func(t *testing.T) {
			u, err := url.Parse(tt.Input)
			if err != nil {
				t.Fatalf("failed to parse URL: %s", err)
			}

			o := main.NormalizeURL(u)
			if o.String() != tt.Output {
				t.Errorf("expected %s but got %s", tt.Output, o)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	StartTestServer(t)

	tests := []struct {
		Target, Pattern string
	}{
		{"ftp://localhost:21021", "\tHEALTHY\t[.0-9]+\tftp://localhost:21021\tsucceed connect\n$"},
		{"ftp://localhost:21022", "\tFAILURE\t[.0-9]+\tftp://localhost:21022\tdial tcp [^ ]+:21022: [^\t]+\n$"},
		{"ftp://hoge:fuga@localhost:21021", "\tHEALTHY\t[.0-9]+\tftp://hoge:xxxxx@localhost:21021\tsucceed connect and login\n$"},
		{"ftp://invalid:user@localhost:21021", "\tFAILURE\t[.0-9]+\tftp://invalid:xxxxx@localhost:21021\t530 Incorrect password, not logged in\n$"},
		{"ftp://hoge@localhost:21021", "\tFAILURE\t[.0-9]+\tftp://hoge@localhost:21021\t553 action aborted, required param missing\n$"},
	}

	for _, tt := range tests {
		t.Run(tt.Target, func(t *testing.T) {
			u, err := url.Parse(tt.Target)
			if err != nil {
				t.Fatalf("failed to parse URL: %s", err)
			}

			buf := &bytes.Buffer{}
			logger := ayd.NewLoggerWithWriter(buf, u)

			main.Check(logger, u)

			if ok, err := regexp.MatchString(tt.Pattern, buf.String()); err != nil {
				t.Errorf("failed to test log: %s", err)
			} else if !ok {
				t.Errorf("unexpected log\nwant pattern:\n%#v\ngot:\n%#v", tt.Pattern, buf.String())
			}
		})
	}
}
