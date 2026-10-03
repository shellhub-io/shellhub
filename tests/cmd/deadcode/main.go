package main

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	workspace := flag.String("workspace", "", "directory holding the shellhub and cloud checkouts (default: this shellhub checkout, with cloud next to its main checkout)")
	requireCloud := flag.Bool("require-cloud", false, "fail instead of skipping when the cloud checkout is missing")
	tagSets := flag.String("tags", "enterprise,docker;enterprise,native", "semicolon-separated build tag sets; code counts as dead only when it is dead under every set")
	base := flag.String("base", "", "git ref each checkout is measured against through its merge base with HEAD, such as origin/master; report only what the change introduces")
	flag.Parse()

	if err := run(context.Background(), *workspace, *base, *requireCloud, strings.Split(*tagSets, ";")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, workspace, baseRef string, requireCloud bool, tagSets []string) error {
	shellhubRoot, cloudRoot, err := checkouts(workspace)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(cloudRoot, "go.mod")); err != nil {
		if requireCloud {
			return fmt.Errorf("cloud checkout not found at %s", cloudRoot)
		}

		fmt.Fprintf(os.Stderr, "skipping: cloud is not checked out next to shellhub (%s), and shellhub alone cannot tell dead code from code only cloud calls\n", cloudRoot)

		return nil
	}

	deadcode, err := deadcodeBinary(ctx, shellhubRoot)
	if err != nil {
		return err
	}

	head, err := analyze(ctx, shellhubRoot, cloudRoot, tagSets, deadcode)
	if err != nil {
		return err
	}

	if baseRef == "" {
		return report(os.Stdout, head, 0)
	}

	baseline, err := os.MkdirTemp("", "deadcode-base-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(baseline) }()

	for name, root := range map[string]string{"shellhub": shellhubRoot, "cloud": cloudRoot} {
		if err := exportMergeBase(ctx, root, baseRef, filepath.Join(baseline, name)); err != nil {
			return err
		}
	}

	base, err := analyze(ctx, filepath.Join(baseline, "shellhub"), filepath.Join(baseline, "cloud"), tagSets, deadcode)
	if err != nil {
		return fmt.Errorf("base: %w", err)
	}

	introduced := head.since(base)

	return report(os.Stdout, introduced, head.count()-introduced.count())
}

func deadcodeBinary(ctx context.Context, shellhubRoot string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "tool", "-n", "deadcode")
	cmd.Dir = filepath.Join(shellhubRoot, "tests")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("building deadcode: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func analyze(ctx context.Context, shellhubRoot, cloudRoot string, tagSets []string, deadcode string) (problems, error) {
	shellhub := repository{
		name:    "shellhub",
		root:    shellhubRoot,
		modules: []string{".", "server", "agent", "gateway", "openapi", "tests"},
	}
	cloud := repository{
		name:    "cloud",
		root:    cloudRoot,
		modules: []string{"."},
	}
	repos := []repository{shellhub, cloud}

	work, err := writeWorkspace(shellhub, repos)
	if err != nil {
		return problems{}, err
	}
	defer func() { _ = os.RemoveAll(filepath.Dir(work)) }()

	dead, err := deadUnderEvery(tagSets, func(tags string) (map[string]finding, error) {
		return collect(ctx, shellhub.root, repos, work, tags, deadcode)
	})
	if err != nil {
		return problems{}, err
	}

	allowed := map[string]string{}
	for _, repo := range repos {
		if err := readAllowlist(repo, allowed); err != nil {
			return problems{}, err
		}
	}

	return findProblems(dead, allowed), nil
}

func exportMergeBase(ctx context.Context, root, ref, dest string) error {
	git := func(args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "git", append([]string{"-c", "safe.directory=" + root, "-C", root}, args...)...) //nolint:gosec // git on a checkout this tool analyses, with the ref from its own flag
	}

	out, err := git("merge-base", "HEAD", ref).Output()
	if err != nil {
		return fmt.Errorf("%s: no merge base between HEAD and %s; a shallow clone needs its full history: %w", root, ref, err)
	}

	archive := git("archive", "--format=tar", strings.TrimSpace(string(out)))

	tree, err := archive.StdoutPipe()
	if err != nil {
		return err
	}

	if err := archive.Start(); err != nil {
		return err
	}

	if err := untar(tree, dest); err != nil {
		_ = archive.Wait()

		return err
	}

	return archive.Wait()
}

func untar(r io.Reader, dest string) error {
	if err := os.MkdirAll(dest, 0o750); err != nil {
		return err
	}

	root, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	archive := tar.NewReader(r)

	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		if err := root.MkdirAll(filepath.Dir(header.Name), 0o750); err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err = root.MkdirAll(header.Name, 0o750)
		case tar.TypeSymlink:
			err = root.Symlink(header.Linkname, header.Name)
		case tar.TypeReg:
			err = writeFile(root, header.Name, archive, header.FileInfo().Mode().Perm())
		}

		if err != nil {
			return err
		}
	}
}

func writeFile(root *os.Root, name string, content io.Reader, mode os.FileMode) error {
	file, err := root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(file, content); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

func checkouts(workspace string) (string, string, error) {
	if workspace != "" {
		return filepath.Join(workspace, "shellhub"), filepath.Join(workspace, "cloud"), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	shellhub, err := findShellHubRoot(wd)
	if err != nil {
		return "", "", err
	}

	return shellhub, filepath.Join(filepath.Dir(mainCheckout(shellhub)), "cloud"), nil
}

func mainCheckout(root string) string {
	data, err := os.ReadFile(filepath.Join(root, ".git")) //nolint:gosec // a linked worktree's .git file names its main checkout
	if err != nil {
		return root
	}

	gitdir, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir: ")
	if !ok {
		return root
	}

	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(root, gitdir)
	}

	sep := string(filepath.Separator)
	checkout, _, found := strings.Cut(filepath.Clean(gitdir), sep+".git"+sep+"worktrees"+sep)
	if !found {
		return root
	}

	return checkout
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

func writeWorkspace(shellhub repository, repos []repository) (string, error) {
	goMod, err := os.ReadFile(filepath.Join(shellhub.root, "go.mod"))
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

	for _, module := range shellhub.modules {
		dir := filepath.Join(shellhub.root, module)

		path, err := modulePath(dir)
		if err != nil {
			return "", err
		}

		fmt.Fprintf(&work, "\nreplace %s v0.0.0 => %s\n", path, dir)
	}

	dir, err := os.MkdirTemp("", "deadcode")
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "go.work")

	return path, os.WriteFile(path, []byte(work.String()), 0o600)
}

func modulePath(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // reads a module of the shellhub checkout
	if err != nil {
		return "", err
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		if path, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(path), nil
		}
	}

	return "", fmt.Errorf("%s/go.mod declares no module", dir)
}

func collect(ctx context.Context, dir string, repos []repository, work, tags, deadcode string) (map[string]finding, error) {
	found := map[string]finding{}

	funcs, err := unreachableFuncs(ctx, dir, repos, work, tags, deadcode)
	if err != nil {
		return nil, err
	}

	decls, err := unusedDeclarations(ctx, dir, repos, work, tags)
	if err != nil {
		return nil, err
	}

	for _, f := range append(funcs, decls...) {
		found[f.key()] = f
	}

	return found, nil
}

func deadUnderEvery(tagSets []string, collect func(tags string) (map[string]finding, error)) (map[string]finding, error) {
	var dead map[string]finding
	for _, tags := range tagSets {
		found, err := collect(tags)
		if err != nil {
			return nil, err
		}

		if dead == nil {
			dead = found

			continue
		}

		for key := range dead {
			if _, ok := found[key]; !ok {
				delete(dead, key)
			}
		}
	}

	return dead, nil
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

func unreachableFuncs(ctx context.Context, dir string, repos []repository, work, tags, deadcode string) ([]finding, error) {
	args := []string{"-test", "-json", "-tags", tags, "-filter", "github.com/shellhub-io/", "github.com/shellhub-io/..."}
	cmd := exec.CommandContext(ctx, deadcode, args...) //nolint:gosec // the binary is this module's pinned deadcode tool, the arguments come from this tool's own flags
	cmd.Dir = dir
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
			file := fn.Position.File
			if !filepath.IsAbs(file) {
				file = filepath.Join(dir, file)
			}

			repo, rel, ok := relate(repos, file)
			if !ok || fn.Generated || skipped(rel) {
				continue
			}

			found = append(found, finding{repo: repo.name, path: rel, line: fn.Position.Line, symbol: fn.Name, kind: "func"})
		}
	}

	return found, nil
}

func unusedDeclarations(ctx context.Context, dir string, repos []repository, work, tags string) ([]finding, error) {
	cfg := &packages.Config{
		Context:    ctx,
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests:      true,
		Dir:        dir,
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
	usedBeyondMembers := map[string]bool{}
	memberTypeRefs := map[string]bool{}
	enumTypeOf := map[string]string{}
	membersOf := map[string][]string{}
	interfacesAt := map[string][]*types.Interface{}
	var mocks []types.Type

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, file := range pkg.Syntax {
			filename := pkg.Fset.File(file.Pos()).Name()
			if strings.Contains(filename, "/mocks/") {
				mocks = append(mocks, declaredTypes(pkg, file)...)
			}

			repo, rel, ok := relate(repos, filename)
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
						if gen.Tok == token.CONST && spec.Type != nil {
							ast.Inspect(spec.Type, func(n ast.Node) bool {
								if ident, ok := n.(*ast.Ident); ok {
									memberTypeRefs[pkg.Fset.Position(ident.Pos()).String()] = true
								}

								return true
							})
						}
					case *ast.TypeSpec:
						names = []*ast.Ident{spec.Name}
					}

					for _, name := range names {
						if !name.IsExported() {
							continue
						}

						pos := pkg.Fset.Position(name.Pos())
						declared[pos.String()] = finding{repo: repo.name, path: rel, line: pos.Line, symbol: name.Name, kind: gen.Tok.String()}

						switch obj := pkg.TypesInfo.Defs[name].(type) {
						case *types.Const:
							if named, ok := obj.Type().(*types.Named); ok && named.Obj().Pkg() == obj.Pkg() {
								typePos := pkg.Fset.Position(named.Obj().Pos()).String()
								enumTypeOf[pos.String()] = typePos
								membersOf[typePos] = append(membersOf[typePos], pos.String())
							}
						case *types.TypeName:
							if iface, ok := obj.Type().Underlying().(*types.Interface); ok && iface.NumMethods() > 0 {
								interfacesAt[pos.String()] = append(interfacesAt[pos.String()], iface)
							}
						}
					}
				}
			}
		}

		for ident, obj := range pkg.TypesInfo.Uses {
			at := pkg.Fset.Position(ident.Pos())
			if !obj.Pos().IsValid() || strings.Contains(at.Filename, "/mocks/") {
				continue
			}

			target := pkg.Fset.Position(obj.Pos()).String()
			used[target] = true
			if !memberTypeRefs[at.String()] {
				usedBeyondMembers[target] = true
			}
		}
	})

	live := func(pos string) bool {
		if used[pos] {
			return true
		}

		if typePos, ok := enumTypeOf[pos]; ok {
			if usedBeyondMembers[typePos] || slices.ContainsFunc(membersOf[typePos], func(member string) bool { return used[member] }) {
				return true
			}
		}

		for _, iface := range interfacesAt[pos] {
			for _, mock := range mocks {
				if types.Implements(mock, iface) || types.Implements(types.NewPointer(mock), iface) {
					return true
				}
			}
		}

		return false
	}

	var found []finding
	for pos, f := range declared {
		if !live(pos) {
			found = append(found, f)
		}
	}

	return found, nil
}

func declaredTypes(pkg *packages.Package, file *ast.File) []types.Type {
	var declared []types.Type
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if obj, ok := pkg.TypesInfo.Defs[typeSpec.Name].(*types.TypeName); ok {
				declared = append(declared, obj.Type())
			}
		}
	}

	return declared
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

		allowed[finding{repo: repo.name, path: fields[0], symbol: fields[1]}.key()] = fmt.Sprintf("%s/%s:%d", repo.name, allowlistName, n)
	}

	return scanner.Err()
}

type problems struct {
	dead  map[string]finding
	stale map[string]string
}

func findProblems(dead map[string]finding, allowed map[string]string) problems {
	found := problems{dead: map[string]finding{}, stale: map[string]string{}}

	for key, f := range dead {
		if _, ok := allowed[key]; !ok {
			found.dead[key] = f
		}
	}

	for key, where := range allowed {
		if _, ok := dead[key]; !ok {
			found.stale[key] = where
		}
	}

	return found
}

func (p problems) since(base problems) problems {
	introduced := problems{dead: map[string]finding{}, stale: map[string]string{}}

	for key, f := range p.dead {
		if _, ok := base.dead[key]; !ok {
			introduced.dead[key] = f
		}
	}

	for key, where := range p.stale {
		if _, ok := base.stale[key]; !ok {
			introduced.stale[key] = where
		}
	}

	return introduced
}

func (p problems) count() int {
	return len(p.dead) + len(p.stale)
}

func report(w io.Writer, found problems, atBase int) error {
	var lines, stale []string

	for _, f := range found.dead {
		lines = append(lines, fmt.Sprintf("%s/%s:%d: unused %s %s", f.repo, f.path, f.line, f.kind, f.symbol))
	}

	for key, where := range found.stale {
		stale = append(stale, fmt.Sprintf("%s: %s is used now, or gone; drop the entry", where, key))
	}

	sort.Strings(lines)
	sort.Strings(stale)

	var out strings.Builder
	for _, line := range append(lines, stale...) {
		out.WriteString(line + "\n")
	}

	failed := found.count() > 0
	if failed {
		fmt.Fprintf(&out, "\nRemove the dead code. If it must stay, add \"<path> <symbol> <reason>\" to that repository's %s.\n", allowlistName)
	} else {
		out.WriteString("no dead code\n")
	}

	if atBase > 0 {
		fmt.Fprintf(&out, "%d already at the base, not reported here\n", atBase)
	}

	if _, err := io.WriteString(w, out.String()); err != nil {
		return err
	}

	if failed {
		return errDeadCode
	}

	return nil
}
