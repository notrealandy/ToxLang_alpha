package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/evaluator"
	"github.com/notrealandy/tox/lexer"
	"github.com/notrealandy/tox/parser"
	"github.com/notrealandy/tox/typechecker"
)

func projectRoot(mainPath string, srcDir string) string {
	abs, _ := filepath.Abs(mainPath)
	idx := strings.LastIndex(abs, srcDir)
	if idx == -1 {
		return filepath.Dir(mainPath)
	}
	return abs[:idx]
}

// Helper to load config
func loadConfig(configPath string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// Recursively load and parse all imports, collecting all statements
func loadAndParseFile(path string, loaded map[string]bool, config map[string]interface{}, allStmts *[]ast.Statement) error {
	if loaded[path] {
		return nil // already loaded
	}
	loaded[path] = true

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", path, err)
	}
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors) > 0 {
		return fmt.Errorf("parser errors in %s: %v", path, p.Errors)
	}

	// --- Recursively load imports ---
	projectPrefix := ""
	if pfx, ok := config["project"].(map[string]interface{})["packagePrefix"].(string); ok {
		projectPrefix = pfx
	}
	srcDirs := config["project"].(map[string]interface{})["sourceDirs"].([]interface{})

	for _, stmt := range program {
		if imp, ok := stmt.(*ast.ImportStatement); ok {
			importPath := imp.Path
			// Strip prefix
			if projectPrefix != "" && strings.HasPrefix(importPath, projectPrefix+".") {
				importPath = strings.TrimPrefix(importPath, projectPrefix+".")
			}
			segments := strings.Split(importPath, ".")
			moduleName := segments[len(segments)-1]
			importDir := filepath.Join(segments...)
			importFile := filepath.Join(importDir, moduleName+".tox")

			found := false
			for _, dir := range srcDirs {
				root := projectRoot(path, dir.(string))
				fullPath := filepath.Join(root, dir.(string), importFile)
				if _, err := os.Stat(fullPath); err == nil {
					var importedStmts []ast.Statement
					err := loadAndParseFile(fullPath, loaded, config, &importedStmts)
					if err != nil {
						return err
					}
					// Prefix function names in the imported file
					for _, istmt := range importedStmts {
						if fn, ok := istmt.(*ast.FunctionStatement); ok {
							fnGlobal := *fn
							fnGlobal.Name = moduleName + "." + fn.Name
							*allStmts = append(*allStmts, &fnGlobal)
							*allStmts = append(*allStmts, fn)
						} else {
							*allStmts = append(*allStmts, istmt)
						}
					}
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("import not found: %s", imp.Path)
			}
		}
	}

	// --- Enforce package statement matches directory structure ---
	var declaredPkg string
	for _, stmt := range program {
		if pkgStmt, ok := stmt.(*ast.PackageStatement); ok {
			declaredPkg = pkgStmt.Name
			break
		}
	}
	// Compute expected package from file path (relative to src)
	srcRoot := ""
	for _, dir := range srcDirs {
		dirStr := dir.(string)
		idx := strings.Index(path, dirStr)
		if idx != -1 {
			srcRoot = path[:idx+len(dirStr)]
			break
		}
	}
	relPath, _ := filepath.Rel(srcRoot, path)
	relPath = strings.TrimSuffix(relPath, ".tox")
	expectedPkg := strings.ReplaceAll(relPath, string(os.PathSeparator), ".")
	expectedPkg = strings.TrimLeft(expectedPkg, ".")
	// Add this line to strip the prefix from declaredPkg for comparison:
	if projectPrefix != "" && strings.HasPrefix(declaredPkg, projectPrefix+".") {
		declaredPkg = strings.TrimPrefix(declaredPkg, projectPrefix+".")
	}
	if declaredPkg != "" {
		// If this is the main file at src/main.tox, allow the prefix as the package
		if expectedPkg == "main" && (declaredPkg == projectPrefix || declaredPkg == "main") {
			// OK
		} else {
			declaredSegments := strings.Split(declaredPkg, ".")
			expectedSegments := strings.Split(expectedPkg, ".")
			if declaredSegments[len(declaredSegments)-1] != expectedSegments[len(expectedSegments)-1] {
				return fmt.Errorf("package name mismatch: file declares '%s', but expected '%s' based on directory", declaredPkg, expectedPkg)
			}
		}
	}
	// Add this file's statements (after imports)
	*allStmts = append(*allStmts, program...)
	return nil
}

func main() {
	// Usage instructions
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Println("Usage: tox run <path>")
		os.Exit(1)
	}

	// Determine the path
	var path string
	if len(os.Args) < 3 || os.Args[2] == "." {
		path = "main.tox"
	} else {
		path = os.Args[2]
	}

	// Load config
	config, err := loadConfig(filepath.Join(filepath.Dir(path), "../tox.config.json"))
	if err != nil {
		fmt.Println("Error loading tox.config.json:", err)
		os.Exit(1)
	}

	// Recursively load all files and collect all statements
	loaded := map[string]bool{}
	var allStmts []ast.Statement
	err = loadAndParseFile(path, loaded, config, &allStmts)
	if err != nil {
		fmt.Println("Import error:", err)
		os.Exit(1)
	}

	// Run typechecker
	errors := typechecker.Check(allStmts)
	if len(errors) > 0 {
		fmt.Println("Type errors:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		os.Exit(1)
	}
	fmt.Println("Program passed type checking ✅\n")

	env := evaluator.NewEnvironment()

	// Evaluate all top-level statements to populate env
	evaluator.Eval(allStmts, env)

	// Now run main if it exists
	if mainFn, ok := env.Get("main"); ok {
		if fnStmt, ok := mainFn.(*ast.FunctionStatement); ok {
			mainEnv := evaluator.NewEnclosedEnvironment(env)
			evaluator.Eval(fnStmt.Body, mainEnv)
		}
	}
}
