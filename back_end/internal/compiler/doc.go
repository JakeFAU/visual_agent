// Package compiler translates validated graph documents into executable Google
// ADK agents.
//
// The package keeps node-specific logic in small NodeCompiler implementations
// and leaves graph-level concerns such as topological ordering, toolbox wiring,
// and branch target resolution to Compiler.
package compiler
