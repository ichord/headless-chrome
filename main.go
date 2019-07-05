package main

import (
	"log"
	"net/http"
	"net/url"
	"fmt"
	"io"
	"strings"
	"bufio"
	"os/exec"
	"context"

	"github.com/yhat/wsutil"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	u, err := launchChrome(ctx)
	defer cancel()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("proxy :9222 to", u)
	err = http.ListenAndServe(":9222", wsutil.NewSingleHostReverseProxy(u))
	if err != nil {
		cancel()
		log.Fatalln(err)
	}
}


type flags []string

func launchChrome(ctx context.Context) (*url.URL, error) {
	args := []string{
	    "--headless",
	    "--disable-gpu",
	    "--disable-software-rasterizer",
	    "--disable-dev-shm-usage",
	    "--hide-scrollbars",
	    "--mute-audio",
	    "--no-default-browser-check=true",
	    "--no-first-run=true",
	    "--disable-background-networking=true",
	    "--enable-features=NetworkService,NetworkServiceInProcess",
	    "--disable-background-timer-throttling=true",
	    "--disable-backgrounding-occluded-windows=true",
	    "--disable-breakpad=true",
	    "--disable-client-side-phishing-detection=true",
	    "--disable-default-apps=true",
	    "--disable-dev-shm-usage=true",
	    "--disable-extensions=true",
	    "--disable-features=site-per-process,TranslateUI,BlinkGenPropertyTrees",
	    "--disable-hang-monitor=true",
	    "--disable-ipc-flooding-protection=true",
	    "--disable-popup-blocking=true",
	    "--disable-prompt-on-repost=true",
	    "--disable-renderer-backgrounding=true",
	    "--disable-sync=true",
	    "--force-color-profile=srgb",
	    "--metrics-recording-only=true",
	    "--safebrowsing-disable-auto-update=true",
	    "--enable-automation=true",
	    "--password-store=basic",
	    "--use-mock-keychain=true",
	}
	args = append(args, "--remote-debugging-port=0")
	args = append(args, "about:blank")
	cmd := exec.CommandContext(ctx, findExecPath(), args...)
	allocateCmdOptions(cmd)

	// We must start the cmd before calling cmd.Wait, as otherwise the two
	// can run into a data race.
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer stderr.Close()
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	u, err := addrFromStderr(stderr)
	if err != nil {
		return nil, err
	}

	return url.Parse(u)
}

func addrFromStderr(rc io.ReadCloser) (string, error) {
	defer rc.Close()
	url := ""
	scanner := bufio.NewScanner(rc)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if s := strings.TrimPrefix(line, "DevTools listening on"); s != line {
			url = strings.TrimSpace(s)
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("chrome stopped too early; stderr:\n%s",
			strings.Join(lines, "\n"))
	}
	return url, nil
}

func findExecPath() string {
	for _, path := range [...]string{
		// Unix-like
		"headless_shell",
		"headless-shell",
		"chromium",
		"chromium-browser",
		"google-chrome",
		"google-chrome-stable",
		"google-chrome-beta",
		"google-chrome-unstable",
		"/usr/bin/google-chrome",

		// Windows
		"chrome",
		"chrome.exe", // in case PATHEXT is misconfigured
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,

		// Mac
		`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
	} {
		found, err := exec.LookPath(path)
		if err == nil {
			return found
		}
	}
	// Fall back to something simple and sensible, to give a useful error
	// message.
	return "google-chrome"
}
