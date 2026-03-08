package h3_test

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type RegistryEntry struct {
	CFile   string `yaml:"c_file"`
	CSuite  string `yaml:"c_suite"`
	CTest   string `yaml:"c_test"`
	GoTest  string `yaml:"go_test"`
	Status  string `yaml:"status"`
	Notes   string `yaml:"notes"`
	SHA256  string `yaml:"sha256"`
}

type Registry struct {
	Version string          `yaml:"version"`
	Entries []RegistryEntry `yaml:"entries"`
}

func TestC_1to1Coverage(t *testing.T) {
	// 1. Walk testdata/c-tests/*.c and extract TEST(suite, name)
	cDir := "testdata/c-tests"
	cFiles, err := filepath.Glob(filepath.Join(cDir, "*.c"))
	if err != nil {
		t.Fatalf("glob c files: %v", err)
	}

	suiteRE := regexp.MustCompile(`SUITE\((\w+)\)\s*{`)
	testRE := regexp.MustCompile(`\s*TEST\((\w+)\)\s*{`)

	// Map: "file::suite::name" → true
	cTests := make(map[string]bool)
	cFileHashes := make(map[string]string)

	for _, cf := range cFiles {
		data, err := os.ReadFile(cf)
		if err != nil {
			t.Fatalf("read %s: %v", cf, err)
		}

		// Compute SHA-256
		h := sha256.New()
		h.Write(data)
		cFileHashes[filepath.Base(cf)] = fmt.Sprintf("%x", h.Sum(nil))

		// Extract suite name
		content := string(data)
		suiteMatch := suiteRE.FindStringSubmatch(content)
		if len(suiteMatch) < 2 {
			t.Logf("warn: no SUITE found in %s", filepath.Base(cf))
			continue
		}
		suite := suiteMatch[1]

		// Extract all test names
		testMatches := testRE.FindAllStringSubmatch(content, -1)
		for _, m := range testMatches {
			testName := m[1]
			key := filepath.Base(cf) + "::" + suite + "::" + testName
			cTests[key] = true
		}
	}

	// 2. Load registry
	regData, err := os.ReadFile("testdata/c_test_registry.yaml")
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var reg Registry
	if err := yaml.Unmarshal(regData, &reg); err != nil {
		t.Fatalf("parse registry: %v", err)
	}

	// 3. Build set of registered C tests
	registered := make(map[string]RegistryEntry)
	for _, entry := range reg.Entries {
		key := entry.CFile + "::" + entry.CSuite + "::" + entry.CTest
		registered[key] = entry
	}

	// 4. Every C test must have a registry entry
	var unregistered []string
	for key := range cTests {
		if _, ok := registered[key]; !ok {
			unregistered = append(unregistered, key)
		}
	}
	sort.Strings(unregistered)
	if len(unregistered) > 0 {
		t.Errorf("C tests without registry entries:\n%s", strings.Join(unregistered, "\n"))
	}

	// 5. Collect all Go test functions from *_test.go files in the root package
	goTests := make(map[string]bool)
	goFiles, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatalf("glob test files: %v", err)
	}
	fset := token.NewFileSet()
	for _, gf := range goFiles {
		// Skip lint_test.go itself to avoid circular issues
		if gf == "lint_test.go" {
			continue
		}
		f, err := parser.ParseFile(fset, gf, nil, 0)
		if err != nil {
			t.Logf("warn: parse %s: %v", gf, err)
			continue
		}
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if strings.HasPrefix(fn.Name.Name, "Test") {
					goTests[fn.Name.Name] = true
				}
			}
		}
	}

	// 6. Every 'covered' entry must have a go_test that exists
	for _, entry := range reg.Entries {
		if entry.Status == "covered" {
			if entry.GoTest == "" {
				t.Errorf("covered entry %s::%s::%s has empty go_test", entry.CFile, entry.CSuite, entry.CTest)
				continue
			}
			if !goTests[entry.GoTest] {
				t.Errorf("covered entry %s::%s::%s references go_test %q which doesn't exist",
					entry.CFile, entry.CSuite, entry.CTest, entry.GoTest)
			}
		}
	}

	// 7. Zero pending entries
	for _, entry := range reg.Entries {
		if entry.Status == "pending" {
			t.Errorf("pending entry found: %s::%s::%s (go_test: %s)",
				entry.CFile, entry.CSuite, entry.CTest, entry.GoTest)
		}
	}

	// 8. Verify SHA-256 of C files matches stored checksums
	for _, entry := range reg.Entries {
		if entry.SHA256 == "" {
			continue
		}
		computedHash := cFileHashes[entry.CFile]
		if computedHash != entry.SHA256 {
			t.Errorf("C file %s SHA-256 mismatch: stored=%s computed=%s",
				entry.CFile, entry.SHA256, computedHash)
		}
	}

	t.Logf("Registry: %d entries, C tests found: %d", len(reg.Entries), len(cTests))
}
