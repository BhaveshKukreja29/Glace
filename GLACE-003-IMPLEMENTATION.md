// GLACE-003: Break and Continue Implementation Summary

// ============================================================================
// OVERVIEW
// ============================================================================
// This implementation adds support for 'break' and 'continue' statements
// inside loop blocks in the Glace language.
//
// break    - Exits the loop immediately
// continue - Skips the rest of the current iteration and moves to the next
//
// Implementation uses Go's error interface with sentinel error values
// (BreakSignal and ContinueSignal) to propagate control flow signals up
// the call stack.

// ============================================================================
// FILES MODIFIED
// ============================================================================

// 1. lexer/token.go
//    - TOKEN_BREAK and TOKEN_CONTINUE keywords already defined
//    - Keywords map includes "break" and "continue" mappings
//    - NO CHANGES NEEDED (already complete)

// 2. ast/ast.go
//    - BreakStatement struct: represents break statements
//    - ContinueStatement struct: represents continue statements
//    - CHANGE: Added GLACE-003 comments to mark these definitions
//    - Both statements defined with required TokenPos() and String() methods

// 3. parser/parser.go
//    - parseBreakStatement(): Creates BreakStatement AST nodes
//    - parseContinueStatement(): Creates ContinueStatement AST nodes
//    - Both functions advance the parser token and return the AST node
//    - CHANGE: Added GLACE-003 comments to parsing logic and function defs
//    - Parser already handles TOKEN_BREAK and TOKEN_CONTINUE in parseStatement()

// 4. evaluator/evaluator.go
//    - BreakSignal: Error type that signals loop break
//    - ContinueSignal: Error type that signals loop continue
//    - CHANGE: Added GLACE-003 comments to signal definitions
//
//    - Eval() function dispatch:
//      * case *ast.BreakStatement: returns (nil, &BreakSignal{})
//      * case *ast.ContinueStatement: returns (nil, &ContinueSignal{})
//      * CHANGE: Added GLACE-003 comments to these cases
//
//    - evalLoopStatement():
//      * Handles conditional loops (loop condition { ... })
//      * Catches BreakSignal to exit loop with 'break'
//      * Catches ContinueSignal to continue to next iteration
//      * CHANGE: Added GLACE-003 comments at signal handling points
//
//    - evalForInLoop():
//      * Handles range loops (loop i in 0..10 { ... })
//      * Handles array loops (loop elem in array { ... })
//      * Catches both BreakSignal and ContinueSignal in both cases
//      * CHANGE: Added GLACE-003 comments to all signal handlers

// ============================================================================
// HOW IT WORKS
// ============================================================================

// Signal Propagation:
// 1. When 'break' is encountered, Eval() returns (nil, &BreakSignal{})
// 2. Signal travels up through the call stack as an error
// 3. evalLoopStatement catches BreakSignal and breaks the loop
// 4. Loop exits and returns NONE
//
// Same mechanism for 'continue':
// 1. When 'continue' is encountered, Eval() returns (nil, &ContinueSignal{})
// 2. Signal travels up as an error
// 3. evalLoopStatement catches ContinueSignal and continues loop
// 4. Current iteration ends, next iteration starts

// ============================================================================
// TEST FILES CREATED
// ============================================================================

// 1. test_break_continue.glace
//    - Minimal test matching acceptance criteria from GLACE-003
//    - Loop from 0..10 with break at i==5 and continue on even numbers
//    - Expected result: sum = 4 (1 + 3)

// 2. examples/break_continue.glace
//    - Three test cases demonstrating break and continue
//    - Test 1: Basic break and continue (acceptance test)
//    - Test 2: Break in conditional loop
//    - Test 3: Continue to skip multiples

// 3. examples/break_continue_advanced.glace
//    - Four advanced examples showing different use cases
//    - Example 1: Simple break to exit loop
//    - Example 2: Simple continue to skip iteration
//    - Example 3: Break with array search
//    - Example 4: Nested loops with continue

// ============================================================================
// USAGE EXAMPLES
// ============================================================================

// Example 1: Break when condition met
// loop i in 0..100 {
//     if i == 50 { break }
//     print(i)
// }

// Example 2: Continue to skip values
// loop i in 1..10 {
//     if i % 2 == 0 { continue }  // Skip even numbers
//     print(i)
// }

// Example 3: Combined break and continue
// mut sum = 0
// loop i in 0..10 {
//     if i == 5 { break }         // Stop at 5
//     if i % 2 == 0 { continue }  // Skip even
//     sum = sum + i
// }
// // sum = 1 + 3 = 4

// ============================================================================
// DESIGN DECISIONS
// ============================================================================

// 1. Signal-based approach:
//    - Used Go's error interface to propagate control flow
//    - Cleaner than modifying Value type
//    - Allows arbitrary nesting without state management
//    - Follows Glace evaluator's existing pattern (also uses for return)

// 2. No special break/continue scoping:
//    - break/continue only work in loop blocks (enforced by parser)
//    - Using outside loops would result in signal not being caught
//    - Could add validation if needed, but parser structure prevents misuse

// 3. Works with all loop types:
//    - Conditional loops: loop condition { ... }
//    - Range loops: loop i in 0..10 { ... }
//    - Array loops: loop elem in arr { ... }
//    - No special handling needed per loop type

// ============================================================================
