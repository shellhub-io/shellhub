package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

const allowlistName = ".deadcode-allow"

var errDeadCode = errors.New("dead code found")

type repository struct {
	name    string
	root    string
	modules []string
}

type finding struct {
	repo   string
	path   string
	line   int
	symbol string
	kind   string
}

func (f finding) key() string {
	return f.repo + "/" + f.path + " " + f.symbol
}

func main() {
	workspace := flag.String("workspace", "", "directory holding the shellhub and cloud checkouts (default: the parent of this shellhub checkout)")
	requireCloud := flag.Bool("require-cloud", false, "fail instead of skipping when the cloud checkout is missing")
	tagSets := flag.String("tags", "enterprise,docker;enterprise,native", "semicolon-separated build tag sets; code counts as dead only when it is dead under every set")
	flag.Parse()

	if err := run(context.Background(), *workspace, *requireCloud, strings.Split(*tagSets, ";")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, workspace string, requireCloud bool, tagSets []string) error {
	if workspace == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		shellhub, err := findShellHubRoot(wd)
		if err != nil {
			return err
		}

		workspace = filepath.Dir(shellhub)
	}

	repos := []repository{
		{name: "shellhub", modules: []string{".", "server", "agent", "gateway", "openapi", "tests"}},
		{name: "cloud", modules: []string{"."}},
	}

	for i := range repos {
		repos[i].root = filepath.Join(workspace, repos[i].name)
	}

	if _, err := os.Stat(filepath.Join(repos[1].root, "go.mod")); err != nil {
		if requireCloud {
			return fmt.Errorf("cloud checkout not found at %s", repos[1].root)
		}

		fmt.Fprintf(os.Stderr, "skipping: cloud is not checked out next to shellhub (%s), and shellhub alone cannot tell dead code from code only cloud calls\n", repos[1].root)

		return nil
	}

	work, err := writeWorkspace(repos)
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(filepath.Dir(work)) }()

	var dead map[string]finding
	for _, tags := range tagSets {
		found, err := collect(ctx, repos, work, tags)
		if err != nil {
			return err
		}

		dead = intersect(dead, found)
	}

	allowed := map[string]string{}
	for _, repo := range repos {
		if err := readAllowlist(repo, allowed); err != nil {
			return err
		}
	}

	return report(dead, allowed)
}

func findShellHubRoot(dir string) (string, error) {
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // walks up from the working directory to find the checkout
		if err == nil && strings.HasPrefix(string(data), "module github.com/shellhub-io/shellhub\n") {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run from inside the shellhub checkout or pass -workspace")
		}

		dir = parent
	}
}

func writeWorkspace(repos []repository) (string, error) {
	goMod, err := os.ReadFile(filepath.Join(repos[0].root, "go.mod"))
	if err != nil {
		return "", err
	}

	version := ""
	for line := range strings.SplitSeq(string(goMod), "\n") {
		if v, ok := strings.CutPrefix(line, "go "); ok {
			version = strings.TrimSpace(v)

			break
		}
	}

	var work strings.Builder
	fmt.Fprintf(&work, "go %s\n\nuse (\n", version)
	for _, repo := range repos {
		for _, module := range repo.modules {
			fmt.Fprintf(&work, "\t%s\n", filepath.Join(repo.root, module))
		}
	}
	work.WriteString(")\n")

	dir, err := os.MkdirTemp("", "deadcode")
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "go.work")

	return path, os.WriteFile(path, []byte(work.String()), 0o600)
}

func collect(ctx context.Context, repos []repository, work, tags string) (map[string]finding, error) {
	found := map[string]finding{}

	funcs, err := unreachableFuncs(ctx, repos, work, tags)
	if err != nil {
		return nil, err
	}

	decls, err := unusedDeclarations(ctx, repos, work, tags)
	if err != nil {
		return nil, err
	}

	for _, f := range append(funcs, decls...) {
		found[f.key()] = f
	}

	return found, nil
}

func intersect(acc, next map[string]finding) map[string]finding {
	if acc == nil {
		return next
	}

	for key := range acc {
		if _, ok := next[key]; !ok {
			delete(acc, key)
		}
	}

	return acc
}

func relate(repos []repository, file string) (repository, string, bool) {
	for _, repo := range repos {
		if rel, err := filepath.Rel(repo.root, file); err == nil && !strings.HasPrefix(rel, "..") {
			return repo, filepath.ToSlash(rel), true
		}
	}

	return repository{}, "", false
}

func skipped(rel string) bool {
	return strings.HasSuffix(rel, "_test.go") || strings.Contains(rel, "/mocks/") || strings.Contains(rel, "node_modules/")
}

func unreachableFuncs(ctx context.Context, repos []repository, work, tags string) ([]finding, error) {
	args := []string{"tool", "deadcode", "-test", "-json", "-tags", tags, "-filter", "github.com/shellhub-io/", "github.com/shellhub-io/..."}
	cmd := exec.CommandContext(ctx, "go", args...) //nolint:gosec // the arguments come from this tool's own flags
	cmd.Dir = repos[0].root
	cmd.Env = append(os.Environ(), "GOWORK="+work, "GOFLAGS=-buildvcs=false")
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("deadcode: %w", err)
	}

	var pkgs []struct {
		Funcs []struct {
			Name     string `json:"Name"`
			Position struct {
				File string `json:"File"`
				Line int    `json:"Line"`
			} `json:"Position"`
			Generated bool `json:"Generated"`
		} `json:"Funcs"`
	}
	if err := json.Unmarshal(out, &pkgs); err != nil {
		return nil, fmt.Errorf("deadcode output: %w", err)
	}

	var found []finding
	for _, pkg := range pkgs {
		for _, fn := range pkg.Funcs {
			repo, rel, ok := relate(repos, fn.Position.File)
			if !ok || fn.Generated || skipped(rel) {
				continue
			}

			found = append(found, finding{repo: repo.name, path: rel, line: fn.Position.Line, symbol: fn.Name, kind: "func"})
		}
	}

	return found, nil
}

func unusedDeclarations(ctx context.Context, repos []repository, work, tags string) ([]finding, error) {
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests:      true,
		Dir:        repos[0].root,
		Env:        append(os.Environ(), "GOWORK="+work, "GOFLAGS=-buildvcs=false"),
		BuildFlags: []string{"-tags=" + tags},
	}

	pkgs, err := packages.Load(cfg, "github.com/shellhub-io/...")
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("packages failed to load")
	}

	declared := map[string]finding{}
	used := map[string]bool{}

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, file := range pkg.Syntax {
			repo, rel, ok := relate(repos, pkg.Fset.File(file.Pos()).Name())
			if !ok || skipped(rel) || ast.IsGenerated(file) {
				continue
			}

			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range gen.Specs {
					var names []*ast.Ident
					switch spec := spec.(type) {
					case *ast.ValueSpec:
						names = spec.Names
					case *ast.TypeSpec:
						names = []*ast.Ident{spec.Name}
					}

					for _, name := range names {
						if !name.IsExported() {
							continue
						}

						pos := pkg.Fset.Position(name.Pos())
						declared[pos.String()] = finding{repo: repo.name, path: rel, line: pos.Line, symbol: name.Name, kind: gen.Tok.String()}
					}
				}
			}
		}

		for ident, obj := range pkg.TypesInfo.Uses {
			if !obj.Pos().IsValid() || strings.Contains(pkg.Fset.Position(ident.Pos()).Filename, "/mocks/") {
				continue
			}

			used[pkg.Fset.Position(obj.Pos()).String()] = true
		}
	})

	var found []finding
	for pos, f := range declared {
		if !used[pos] {
			found = append(found, f)
		}
	}

	return found, nil
}

func readAllowlist(repo repository, allowed map[string]string) error {
	file, err := os.Open(filepath.Join(repo.root, allowlistName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return fmt.Errorf("%s/%s:%d: want \"<path> <symbol> <reason>\", got %q", repo.name, allowlistName, n, line)
		}

		allowed[repo.name+"/"+fields[0]+" "+fields[1]] = fmt.Sprintf("%s/%s:%d", repo.name, allowlistName, n)
	}

	return scanner.Err()
}

func report(dead map[string]finding, allowed map[string]string) error {
	var lines, stale []string

	for key, f := range dead {
		if _, ok := allowed[key]; ok {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s/%s:%d: unused %s %s", f.repo, f.path, f.line, f.kind, f.symbol))
	}

	for key, where := range allowed {
		if _, ok := dead[key]; !ok {
			stale = append(stale, fmt.Sprintf("%s: %s is used now, or gone; drop the entry", where, key))
		}
	}

	sort.Strings(lines)
	sort.Strings(stale)

	var out strings.Builder
	for _, line := range append(lines, stale...) {
		out.WriteString(line + "\n")
	}

	failed := len(lines) > 0 || len(stale) > 0
	if failed {
		fmt.Fprintf(&out, "\nRemove the dead code. If it must stay, add \"<path> <symbol> <reason>\" to that repository's %s.\n", allowlistName)
	} else {
		out.WriteString("no dead code\n")
	}

	if _, err := os.Stdout.WriteString(out.String()); err != nil {
		return err
	}

	if failed {
		return errDeadCode
	}

	return nil
}
