// Copyright 2023 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
Package ingress provides functionality for registering services.

The ingress package allows you to register wayback services.

To use the ingress package, import the package and its dependencies,
such as register, as shown in the following example:

	package main

	import (
	        "github.com/wabarc/wayback/ingress"
	        "github.com/wabarc/wayback/publish"
	        "github.com/wabarc/wayback/service"
	        _ "github.com/wabarc/wayback/ingress/register"
	)

	func main() {
	        // Initialize the publish service with configuration options and a context.
	        opts := &config.Options{}
	        ctx := context.Background()
	        pub := publish.New(ctx, opts)
	        go pub.Start()
	        defer pub.Stop()

	        ingress.Init(opts)

	        // Use the publish service to publish data.
	        // ...

	        // Initialize services with configuration options and a context.
	        err := service.Serve(ctx, service.Options{})
	        // ...
	}
*/
package ingress // import "github.com/wabarc/wayback/ingress"
