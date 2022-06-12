// Copyright 2022 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package planner // import "github.com/wabarc/wayback/planner"

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/cdproto/css"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/wabarc/logger"
	"github.com/wabarc/starter/installer"
	"github.com/wabarc/wayback/config"
	"github.com/wabarc/wayback/errors"
)

var (
	remoteDebuggingPort = "9223"
	remoteDebuggingAddr = "localhost"

	recaptchaIframe = "/recaptcha/api2/anchor"
	dialogSelector  = "rc-anchor-container"
)

type today struct {
	workspace   string
	userDataDir string
}

func (t today) init(ch chan bool) error {
	starter := &installer.Starter{
		Home: t.workspace,
	}
	err := starter.Install()
	if err != nil {
		return errors.Wrap(err, "install starter failed")
	}
	ch <- true

	// Installed starter's executable binary path (chromium)
	command := starter.Command()

	opts := []string{
		"-remote-debugging-port=" + remoteDebuggingPort,
		"-workspace=" + t.workspace,
	}
	if os.Getenv("HEADLESS") == "false" {
		opts = append(opts, "-desktop")
	}
	cmd := exec.Command(command, opts...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "run starter failed")
	}

	go readOutput(stdout)

	// Wait for the process to be finished.
	// Don't care about this error in any scenario.
	_ = cmd.Wait()

	return nil
}

// run handles to regularly update the 'ARCHIVE_COOKIE' environment
func (t today) run(ctx context.Context) error {
	// Due to issue#505 (https://github.com/chromedp/chromedp/issues/505),
	// chrome restricts the host must be IP or localhost, we should rewrite the url.
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/json/version", remoteDebuggingAddr, remoteDebuggingPort), nil)
	if err != nil {
		return errors.Wrap(err, "new request chromium failed")
	}

	req.Host = net.JoinHostPort(remoteDebuggingAddr, remoteDebuggingPort)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request remote chromium failed")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return errors.Wrap(err, "decode socket response failed")
	}

	uri := result["webSocketDebuggerUrl"].(string)
	uri = strings.Replace(uri, "localhost", remoteDebuggingAddr, 1)
	ctx, cancel := chromedp.NewRemoteAllocator(ctx, uri)
	defer cancel()

	opts := chromedp.WithErrorf(log.Printf)
	if config.Opts.HasDebugMode() {
		opts = chromedp.WithDebugf(log.Printf)
	}
	ctx, cancel = chromedp.NewContext(ctx, opts)
	defer cancel()

	// Get the archive.today's final URL.
	// uri = t.resolve("https://archive.ph")
	uri = "http://archive.ph"
	// "archive.today",
	// "archive.is",
	// "archive.li",
	// "archive.vn",
	// "archive.fo",
	// "archive.md",
	// "archive.ph",
	// "archiveiya74codqgiixo33q62qlrqtkgmcitqx5u2oeqnmn5bpcbiyd.onion",
	// err = chromedp.Run(ctx, network.DeleteCookies("cf_clearance"), chromedp.Navigate(uri))
	err = chromedp.Run(ctx, chromedp.Navigate(uri))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("open %s failed", uri))
	}
	logger.Debug("open %s successfully", uri)

	ok := false
	script := `() => {
try{
  const input = document.getElementById('url');
  if (input !== null) {
    input.value = 'https://example.com';
    document.querySelector('#submiturl input[type=submit]').click();
    return true;
  }
  const recaptcha = document.getElementById('g-recaptcha');
  if (recaptcha !== null) {
    return true;
  }
}catch(_){};
return false;
}`
	err = chromedp.Run(ctx,
		dom.Enable(),
		chromedp.Navigate(uri),
		chromedp.PollFunction(script, &ok, chromedp.WithPollingTimeout(10*time.Second)),
	)
	if err != nil {
		return errors.Wrap(err, "trigger recaptcha failed")
	}
	logger.Debug("trigger recaptcha result: %t", ok)

	iframe := `iframe[title="reCAPTCHA"]`
	err = chromedp.Run(ctx, chromedp.WaitVisible(iframe, chromedp.ByQuery))
	if err != nil {
		return errors.Wrap(err, "wait elements failed")
	}

	ictx := getIframeContext(ctx, recaptchaIframe)
	script = fmt.Sprintf(`document.getElementById("%s");`, dialogSelector)
	var b []byte
	err = chromedp.Run(
		ictx,
		chromedp.WaitReady(dialogSelector, chromedp.ByID),
		chromedp.Evaluate(script, &b),
	)
	if err != nil {
		return errors.Wrap(err, "dialog recaptcha failed")
	}

	// paste to console `monitorEvents(window, 'click')` to get the position
	// iframe := `iframe[title="reCAPTCHA"]`
	err = chromedp.Run(ctx,
		dom.Enable(),
		css.Enable(),
		page.Enable(),
		// chromedp.WaitVisible(iframe, chromedp.ByQuery),
		// chromedp.Click(`.g-recaptcha`, chromedp.ByQuery),
		// chromedp.MouseClickXY(300, 290), // dialog reCAPTCHA
		chromedp.MouseClickXY(333, 333), // dialog reCAPTCHA
		chromedp.Sleep(3*time.Second),
		chromedp.MouseClickXY(400, 600), // switch to audio mode
		// chromedp.MouseClickXY(480, 600), // switch to audio mode
		// chromedp.MouseClickXY(430, 560), // switch to audio mode
		// chromedp.MouseClickXY(360, 350), // focus on response input
		chromedp.Sleep(5*time.Second),
		// chromedp.MouseClickXY(395, 535), // switch back to image mode
		// chromedp.MouseClickXY(478, 541), // click download audio
	)
	if err != nil {
		return errors.Wrap(err, "click elements failed")
	}
	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.FullScreenshot(&buf, 100),
	)
	if err := ioutil.WriteFile("fullScreenshot.png", buf, 0o600); err != nil {
		return errors.Wrap(err, "screenshot failed")
	}

	var cookies []*network.Cookie
	err = chromedp.Run(ctx,
		css.Enable(),
		dom.Enable(),
		page.Enable(),
		network.Enable(),
		// chromedp.WaitReady(`link[rel="canonical"]`, chromedp.ByQuery), // <link rel="canonical" href="http://archive.today/">
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			if cookies, err = network.GetAllCookies().Do(ctx); err != nil {
				return err
			}
			return nil
		}),
	)
	if err != nil {
		return errors.Wrap(err, "get cookies failed")
	}

	for _, cookie := range cookies {
		if strings.HasPrefix(cookie.Name, "cf_clearance") {
			pair := fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
			os.Setenv("ARCHIVE_COOKIE", pair)
			logger.Debug("`ARCHIVE_COOKIE` environment successfully set: %s", pair)
		}
	}

	return nil
}

func (t today) resolve(u string) string {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, // nolint:gosec,goimports
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
			MinVersion:               tls.VersionTLS11,
			MaxVersion:               tls.VersionTLS11,
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Head(u)
	if err != nil {
		return u
	}
	defer resp.Body.Close()

	return resp.Request.URL.String()
}

// copied from https://github.com/chromedp/chromedp/issues/72#issuecomment-570791861
func getIframeContext(ctx context.Context, uriPart string) context.Context {
	targets, _ := chromedp.Targets(ctx)
	var tgt *target.Info
	for _, t := range targets {
		logger.Debug("%s | %s | %s | %s", t.Title, t.Type, t.URL, t.TargetID)
		if t.Type == "iframe" && strings.Contains(t.URL, uriPart) {
			tgt = t
		}
	}
	if tgt != nil {
		ictx, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(tgt.TargetID))
		return ictx
	}
	return nil
}

func readOutput(rc io.ReadCloser) {
	for {
		out := make([]byte, 1024)
		_, err := rc.Read(out)
		logger.Debug(string(out))
		if err != nil {
			break
		}
	}
}
