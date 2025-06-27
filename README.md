# ToxLang

A toy programming language with basic features.

## Supported Syntax

### Comments
```tox
// This is a comment
```

### Variable Declaration

```tox
let x int >> 5  
let y string >> "hello"  
let z bool >> true
```

### Arithmetic Expressions

```tox
let sum int >> 1 + 2 * 3  
let diff int >> 10 - 4  
let prod int >> 2 * 3  
let quot int >> 8 / 2  
let mod int >> 7 % 3
```

### Functions

```tox
fnc main () >> void {  
    let x int >> 2  
    test()  
    log(x)  
}

fnc test () >> void {  
    let a int >> 11  
    log(a)  
    return nil  
}
```

- Functions are declared with: fnc name () >> return_type { ... }
- Only void and no-parameter functions are supported for now.
- Use return value for non-void, return nil for void.

### Function Calls

- Call functions by name: test()
- Function calls can be used as statements or in assignments.

### Logging

```tox
log(x)  
log("Hello, world!")
```

### Return Statements

```tox
return 42  
return nil  // for void functions
```

### Conditionals: if, elif, else

ToxLang supports `if`, `elif`, and `else` statements for control flow.

```tox
if x > 1 && x < 10 {
    log("x is greater than 1 and less than 10")
} elif x == 1 {
    log("x is exactly 1")
} else {
    log("x is not in the range")
}
```

- The `if` block runs if its condition is true.
- Any number of `elif` blocks can follow, each with their own condition.
- The `else` block is optional and runs if none of the previous conditions matched.
- You can use all supported operators (`==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`) and parentheses in conditions.

**Examples:**

```tox
if x > 100 {
    log("big!")
} elif x > 10 && x < 100 {
    log("medium!")
} else {
    log("small or negative!")
}
```

### Types

- int
- string
- bool
- void (for function return type)

### Notes

- All variables are global.
- No function parameters or argument passing yet.
- No conditionals, loops, or arrays yet.
- Only integer arithmetic and function calls are supported in expressions.

---

## Example

```tox
fnc main () >> void {  
    let x int >> 2  
    test()  
    log(x)  
}

fnc test () >> void {  
    let a int >> 11  
    log(a)  
    return nil  
}
```

---

## TODO

See TODO.md for planned features.
