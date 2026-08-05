// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Command praetorian runs the wrap/unwrap HTTP service defined by the
// praetorian package. Configuration is read entirely from the environment:
// PRAETORIAN_CONFIG supplies the root key material (see
// [praetorian.NewConfig]) and PORT selects the listen port, defaulting to
// 3000 if unset. The process serves until it receives SIGINT, at which
// point it attempts a graceful shutdown before exiting.
package main

import (
	"fmt"
	"os"

	"github.com/karlbateman/praetorian"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	c, err := praetorian.NewConfig()
	if err != nil {
		return err
	}
	ks, err := praetorian.NewKeystore(c)
	if err != nil {
		return err
	}
	return praetorian.NewServer(ks).Start()
}
