package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/fsnotify/fsnotify"
)

// NginxController manages the configuration and operation of NGINX.
type NginxController struct {
	rootDir       string
	templatesDir  string
	gatewayConfig *GatewayConfig
	process       *os.Process
}

// generateConfigs generates the NGINX configuration files.
func (nc *NginxController) generateConfigs() {
	if err := os.MkdirAll("/etc/nginx", 0o755); err != nil {
		log.Fatalf("Failed to create nginx directory: %v", err)
	}

	// Recursively copy all template files maintaining the directory structure
	err := filepath.Walk(nc.templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "~") {
			return nil
		}

		relativePath, err := filepath.Rel(nc.templatesDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(nc.rootDir, relativePath)

		if info.IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
		} else {
			nc.generateConfig(path, destPath)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to process templates: %v", err)
	}
}

// generateConfig processes a template file and writes the result to the destination path.
func (nc *NginxController) generateConfig(src, dst string) {
	var err error
	tmpl := template.New(filepath.Base(src))
	tmpl, err = tmpl.Funcs(template.FuncMap{
		"args": templateArgs,
		"set_upstream": func(host string, port int) (string, error) {
			return buildUpstreamConfig(tmpl, host, port)
		},
	}).ParseFiles(src)
	if err != nil {
		log.Fatalf("Failed to parse template file %s: %v", src, err)
	}

	output := &bytes.Buffer{}
	err = tmpl.Execute(output, map[string]interface{}{
		"Config": nc.gatewayConfig,
	})
	if err != nil {
		log.Fatalf("Failed to execute template %s: %v", src, err)
	}

	err = os.WriteFile(dst, output.Bytes(), 0o600)
	if err != nil {
		log.Fatalf("Failed to write config file %s: %v", dst, err)
	}
}

// watchConfigTemplates watches for changes in the template directory and regenerates the configs on changes.
func (nc *NginxController) watchConfigTemplates() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("watching config templates")

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.HasSuffix(event.Name, "~") {
					break
				}
				if event.Has(fsnotify.Write) {
					fmt.Println("GatewayConfig file modified:", event.Name)
					nc.generateConfigs()
					nc.reload()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watcher error:", err)
			}
		}
	}()

	err = filepath.WalkDir(nc.templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Println("ERROR: Failed to add directory to watcher:", err)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

// reload sends a SIGHUP signal to the NGINX process to reload the configuration.
func (nc *NginxController) reload() {
	if nc.process == nil {
		return
	}

	log.Println("Reloading nginx process")

	if err := nc.process.Signal(syscall.SIGHUP); err != nil {
		log.Fatal("Failed to reload NGINX")
	}
}

// start initializes and starts the NGINX process.
func (nc *NginxController) start() {
	fmt.Println("Starting NGINX")

	cmd := exec.Command(
		"/usr/local/openresty/nginx/sbin/nginx",
		"-c", "/etc/nginx/nginx.conf",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}

	// Goroutine to handle termination signals and gracefully stop NGINX
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Fatal(err)
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	nc.process = cmd.Process

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

// templateArgs function to be used for argument handling in templates.
func templateArgs(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("misaligned args")
	}

	argsMap := make(map[string]any)
	for i := 0; i < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); ok {
			argsMap[key] = pairs[i+1]
		} else {
			return nil, fmt.Errorf("key must be a string, got %T", pairs[i])
		}
	}

	return argsMap, nil
}

func buildUpstreamConfig(tmpl *template.Template, host string, port int) (string, error) {
	output := &bytes.Buffer{}
	err := tmpl.Lookup("UPSTREAM_CONFIG").Execute(
		output,
		map[string]interface{}{
			"Host": host,
			"Port": port,
		},
	)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
