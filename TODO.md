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
- [x] default visibility of function/variable private
- [x] add public function/variable by typing `pub` before it `pub let x int >> 1` or `pub fnc helloTox () >> void {}`

## Loops
- [x] Add support for `while` or `for` loops.
- [ ] Range based for loops

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
- [x] Add more built-in functions (e.g., `len`, `input`, etc.).

## Comments
- [x] Allow comments in source code (e.g., lines starting with `//`).

## Break and Continue
- [x] break
- [x] continue

## Packages and imports
- [x] Packages
- [x] Imports

## Structs
- [x] Structs
- [x] Support for field assignments (updating a field in an already-created struct)
- [x] Methods on structs
- [x] More detailed validation (e.g. checking that all fields are provided or no extra fields exist)

## Lambadas/maps
- [ ] Lambadas (anonymous functions)
- [x] maps

## Compound operators
- [ ] +=
- [ ] -=
- [ ] ++
- [ ] --

## String Interpolation
- [x] "Hello, <%name%>"

## Pattern Matching
- [ ] 

## Strings
- [x] multiline strings


## Memory management
Go manages memory automatically (garbage collected), but if you want manual control in Tox, you could expose:

- [x] make([]byte, size) for buffer allocation

- [x] copy for copying slices

- [x] len and cap for slice introspection

Possibly you don’t need malloc/free bindings because Go handles that internally.

## File I/O
Go’s os and io packages provide all you need:

- [x] os.Open, os.Create, os.Remove (file operations)

- [x] Read, Write on files (implement as Go functions exposed to Tox)

- [x] os.Stat for file metadata

- [x] bufio for buffered reading/writing

## Console I/O

- [x] fmt.Println, fmt.Printf for output (wrap as Tox built-ins)

- [x] bufio.Reader(os.Stdin) for input

## String and byte manipulation

- [x] strings package functions (split, trim, etc.)

- [x] bytes package (buffer manipulation)

## Time & Date

- [x] time.Now(), time.Sleep()