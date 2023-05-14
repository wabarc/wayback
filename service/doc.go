// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package service implements the common utils function for daemon services.

To use the service package, import the package and its dependencies, as shown in the following example:

	package main

	import (
	        _ "github.com/wabarc/wayback/ingress"
	        "github.com/wabarc/wayback/service"
	)

	func main() {
	        // Initialize services with configuration options and a context.
	        opts := service.Options{}
	        ctx := context.Background()
	        err := service.Serve(ctx, opts)
	        // ...
	}
*/
package service // import "github.com/wabarc/wayback/service"
