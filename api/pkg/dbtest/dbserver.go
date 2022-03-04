package dbtest

// mgo - MongoDB driver for Go

// Copyright (c) 2010-2013 - Gustavo Niemeyer <gustavo@niemeyer.net>

// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEV

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/shellhub-io/shellhub/pkg/dockerutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gopkg.in/tomb.v2"
)

func init() {
	cmd := exec.Command("/bin/sh", "-c", "docker info")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "---- Failed to initialize dbtest:\n")
		fmt.Fprint(os.Stderr, out.String())
		panic("Docker is not installed or is not running properly")
	}
}

// DBServer controls a MongoDB server process to be used within test suites.
//
// The test server is started when Client is called the first time and should
// remain running for the duration of all tests, with the Wipe method being
// called between tests (before each of them) to clear stored data. After all tests
// are done, the Stop method should be called to stop the test server.
type DBServer struct {
	Ctx     context.Context
	timeout time.Duration
	client  *mongo.Client
	output  bytes.Buffer
	server  *exec.Cmd
	host    string
	network string
	tomb    tomb.Tomb
}

func (dbs *DBServer) SetTimeout(timeout int) {
	dbs.timeout = time.Duration(timeout)
}

func (dbs *DBServer) start() {
	if dbs.server != nil {
		panic("DBServer already started")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("unable to listen on a local address: " + err.Error())
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		panic("Type assertion failed")
	}

	l.Close()

	dbs.network = "host" // Use same network as docker host
	dbs.host = addr.String()

	if dockerutils.IsRunningInDocker() {
		containerID, err := dockerutils.CurrentContainerID()
		if err != nil {
			panic("failed to get current container id: " + err.Error())
		}

		if containerID != "" {
			// If tests are running in a docker container use the same container network
			dbs.network = fmt.Sprintf("container:%s", containerID)
		}
	}

	args := []string{
		"run", "--rm", fmt.Sprintf("--net=%s", dbs.network), "mongo:4.4.8",
		"--replSet", "rs",
		"--bind_ip", "127.0.0.1",
		"--port", strconv.Itoa(addr.Port),
	}
	dbs.tomb = tomb.Tomb{}
	dbs.server = exec.Command("docker", args...)
	dbs.server.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
	dbs.server.Stdout = &dbs.output
	dbs.server.Stderr = &dbs.output
	err = dbs.server.Start()
	if err != nil {
		// print error to facilitate troubleshooting as the panic will be caught in a panic handler
		fmt.Fprintf(os.Stderr, "mongod failed to start: %v\n", err)
		panic(err)
	}
	dbs.tomb.Go(dbs.monitor)
	dbs.Wipe()
}

func (dbs *DBServer) monitor() error {
	if _, err := dbs.server.Process.Wait(); err != nil {
		return err
	}

	if dbs.tomb.Alive() {
		// Present some debugging information.
		fmt.Fprintf(os.Stderr, "---- mongod container died unexpectedly:\n")
		fmt.Fprintf(os.Stderr, "%s", dbs.output.Bytes())
		fmt.Fprintf(os.Stderr, "---- mongod containers running right now:\n")

		cmd := exec.Command("/bin/sh", "-c", "docker ps --filter ancestor=mongo")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "----------------------------------------\n")

		panic("mongod container died unexpectedly")
	}

	return nil
}

// Stop stops the test server process, if it is running.
//
// It's okay to call Stop multiple times. After the test server is
// stopped it cannot be restarted.
//
// All database clients must be closed before or while the Stop method
// is running. Otherwise Stop will panic after a timeout informing that
// there is a client leak.
func (dbs *DBServer) Stop() {
	if dbs.client != nil {
		if err := dbs.client.Disconnect(dbs.Ctx); err != nil {
			panic("fail to disconnect the database")
		}

		dbs.client = nil
	}

	if dbs.server != nil { //nolint:nestif
		dbs.tomb.Kill(nil)

		// Windows doesn't support Interrupt
		if runtime.GOOS == "windows" {
			if err := dbs.server.Process.Signal(os.Kill); err != nil {
				panic("fail to send os.Kill to the server")
			}
		} else {
			if err := dbs.server.Process.Signal(os.Interrupt); err != nil {
				panic("fail to send os.Interrupt to the server")
			}
		}

		select {
		case <-dbs.tomb.Dead():
		case <-time.After(5 * time.Second):
			panic("timeout waiting for mongod process to die")
		}
		dbs.server = nil
	}
}

// Client returns a new client to the server. The returned client
// must be disconnected after the tests are finished.
//
// The first call to Client will start the DBServer.
func (dbs *DBServer) Client() *mongo.Client {
	if dbs.server == nil {
		dbs.start()
	}

	if dbs.client != nil {
		return dbs.client
	}

	var err error

	if dbs.timeout == 0 {
		dbs.timeout = 8 * time.Second
	}

	// Wait for mongodb to be available
	ticker := time.NewTicker(time.Second)
ticker:
	for {
		select {
		case <-time.After(dbs.timeout):
			panic("mongodb connection timeout")
		case <-ticker.C:
			if _, err := net.Dial("tcp", dbs.host); err != nil {
				continue
			}

			break ticker

		}
	}

	args := []string{
		"run", "--rm", fmt.Sprintf("--net=%s", dbs.network), "mongo:4.4.8",
		"mongo",
		"--host", dbs.host,
		"--eval", "rs.initiate()",
		"--quiet",
	}

	// Initiates mongodb replica set before anything else
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", out)
		panic(err)
	}

	clientOptions := options.Client().ApplyURI("mongodb://" + dbs.host + "/test")
	dbs.Ctx = context.Background()

	dbs.client, err = mongo.Connect(dbs.Ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	if dbs.client == nil {
		panic("cant connect")
	}

	// Verify that the server is accepting connections
	if err := dbs.client.Ping(dbs.Ctx, nil); err != nil {
		panic(err)
	}

	return dbs.client
}

func (dbs *DBServer) CTX() context.Context {
	return dbs.Ctx
}

// Wipe drops all created databases and their data.
func (dbs *DBServer) Wipe() {
	if dbs.server == nil || dbs.client == nil {
		return
	}
	client := dbs.Client()
	names, err := client.ListDatabaseNames(dbs.Ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		switch name {
		case "admin", "local", "config":
		default:
			err = dbs.client.Database(name).Drop(dbs.Ctx)
			if err != nil {
				panic(err)
			}
		}
	}
}
