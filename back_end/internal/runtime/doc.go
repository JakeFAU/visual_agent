// Package runtime provides thin adapters around Google ADK execution backends.
//
// The rest of the application compiles graphs into agents; this package is
// responsible for turning those agents into runnable sessions against either
// in-memory or Vertex-backed session services.
package runtime
