// Package server exposes the HTTP API used by the Visual Agent frontend.
//
// It is intentionally small: validate and persist graphs, compile them on
// demand, and stream execution events back to the browser over SSE.
package server
