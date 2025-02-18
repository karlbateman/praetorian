// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
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
