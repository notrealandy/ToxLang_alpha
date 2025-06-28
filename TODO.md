# TODO List

## Function Parameters and Arguments
- [x] Parse parameters in function definitions.
- [x] Parse arguments in function calls.
- [x] Bind arguments to parameters in a new local environment when calling a function.

## Local Variable Scoping
- [x] Implement local scope for variables inside functions (so variables in one function don’t leak into others or global scope).

## Variable Assignment (not just declaration)
- [x] Support updating the value of an existing variable (e.g., `x >> 10`).

## If Statements / Conditionals
- [x] Add support for `if` and `else` blocks for control flow.

## Switch Statements
- [ ] Add support fot `switch` and `case`

## visibility support acorss files
- [ ] default visibility of function/variable private
- [ ] add public function/variable by typing `pub` before it `pub let x int >> 1` or `pub fnc helloTox () >> void {}`

## Loops
- [x] Add support for `while` or `for` loops.

## Boolean Operators
- [x] Support logical operators: `&&`, `||`, `!`.

## Arrays or Lists
- [x] Add support for array types, literals, and indexing.
- [x] Add Array mutation `xs[0] >> v`
- [x] Add Array length `len(xs)`
- [x] Add Array Slices/Subarrays

## Error Handling
- [ ] Improve error messages for invalid syntax, type errors, and runtime errors.

## Standard Library Functions
- [ ] Add more built-in functions (e.g., `len`, `input`, etc.).

## Comments
- [x] Allow comments in source code (e.g., lines starting with `//`).

## Break and Continue
- [ ] break
- [ ] continue

## Packages and imports
- [ ] Packages
- [ ] Imports

## Structs
- [ ] Structs

## Lambadas/maps
- [ ] Lambadas (anonymous functions)
- [ ] maps

