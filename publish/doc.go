// Copyright 2020 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

/*
The publish package provides a publishing service and requires initialization by the caller.

To use the publish package, import the package and its dependencies, such as register, as shown in the following example:

	package main

	import (
	        _ "github.com/wabarc/wayback/ingress"
	        "github.com/wabarc/wayback/publish"
	)

	func main() {
	        // Initialize the publish service with configuration options and a context.
	        opts := &config.Options{}
	        ctx := context.Background()
	        pub := publish.New(ctx, opts)
	        go pub.Start()
	        defer pub.Stop()

	        // Use the publish service to publish data.
	        // ...
	}
*/
package publish // import "github.com/wabarc/wayback/publish"
