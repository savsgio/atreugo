/*
Package atreugo is a high performance and extensible micro web framework with zero memory allocations in hot paths

It's build on top of fasthttp and provides the following features:

    * Optimized for speed. Easily handles more than 100K qps and more than 1M
        concurrent keep-alive connections on modern hardware.
    * Optimized for low memory usage.
    * Easy 'Connection: Upgrade' support via RequestCtx.Hijack.
    * Server provides the following anti-DoS limits:

        * The number of concurrent connections.
        * The number of concurrent connections per client IP.
        * The number of requests per connection.
        * Request read timeout.
        * Response write timeout.
        * Maximum request header size.
        * Maximum request body size.
        * Maximum request execution time.
        * Maximum keep-alive connection lifetime.
        * Early filtering out non-GET requests.

    * A lot of additional useful info is exposed to request handler:

        * Server and client address.
        * Per-request logger.
        * Unique request id.
        * Request start time.
        * Connection start time.
        * Request sequence number for the current connection.

    * Middlewares support:

        * Before view execution.
        * After view execution.

    * Easy routing:

        * Path parameters (mandatories and optionals).
        * Views with timeout.
        * Group paths and middlewares.
        * Static files.
        * Serve one file like pdf, etc.
        * Filters (middlewares) to specific views.
        * net/http handlers support.
        * fasthttp handlers support

    * Common responses (also you could use your own responses):

        * JSON
        * HTTP
        * Text
        * Raw
        * File
        * Redirect

*/
package atreugo
