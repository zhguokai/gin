package gin

import "github.com/valyala/fasthttp"

// Handler makes the router implement the fasthttp.ListenAndServe interface.
func (engine *Engine) Handler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	if root := engine.trees.get(method); root != nil {
		f, tsr := root.getValue(path)

		if f != nil {
			f(ctx, ps)
			return
		} else if method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && engine.RedirectTrailingSlash {
				var uri string
				if len(path) > 1 && path[len(path)-1] == '/' {
					uri = path[:len(path)-1]
				} else {
					uri = path + "/"
				}
				ctx.Redirect(uri, code)
				return
			}

			// Try to fix the request path
			if engine.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					engine.RedirectTrailingSlash,
				)
				if found {
					uri := string(fixedPath)
					ctx.Redirect(uri, code)
					return
				}
			}
		}
	}

	if method == "OPTIONS" {
		// Handle OPTIONS requests
		if engine.HandleOPTIONS {
			if allow := engine.allowed(path, method); len(allow) > 0 {
				ctx.Response.Header.Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if engine.HandleMethodNotAllowed {
			if allow := engine.allowed(path, method); len(allow) > 0 {
				ctx.Response.Header.Set("Allow", allow)
				if engine.MethodNotAllowed != nil {
					engine.MethodNotAllowed(ctx)
				} else {
					ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
					ctx.SetContentTypeBytes(defaultContentType)
					ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
				}
				return
			}
		}
	}

	// Handle 404
	if engine.NotFound != nil {
		engine.NotFound(ctx)
	} else {
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
			fasthttp.StatusNotFound)
	}

}
