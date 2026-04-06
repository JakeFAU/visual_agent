// Package compiler translates validated graph documents into executable Google
// ADK agents.
//
// The package keeps node-specific logic in small NodeCompiler implementations
// and leaves graph-level concerns such as toolbox wiring, output aliasing, and
// control-flow orchestration to Compiler.
package compiler
