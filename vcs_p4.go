package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/x-motemen/ghq/cmdutil"
	"github.com/x-motemen/ghq/logger"
)

/*
Perforce (p4) workspaces are identified by URLs of the form

	p4://[user@]host[:port]/depot/path       (P4PORT host:port, default port 1666)
	p4+ssl://[user@]host[:port]/depot/path   (P4PORT ssl:host:port)

"ghq get" creates a client workspace whose root is the local directory and
whose view maps //depot/path/... into it, writes the connection settings to
.p4config in the workspace root, and runs "p4 sync".
*/

const (
	p4ConfigName  = ".p4config"
	p4DefaultPort = "1666"
)

var (
	errP4Branch        = errors.New("--branch option (or @<revision> in the URL) is unavailable for Perforce; specify the depot path of the branch in the URL")
	errP4RootUnchanged = errors.New("the root of the p4 client was not changed")
)

// p4Config holds the connection settings recorded in a workspace's .p4config.
type p4Config struct {
	port, user, client string
}

// args returns the p4 global options selecting the server, user and client,
// so that p4 never falls back to P4PORT/P4CLIENT from the environment.
func (c *p4Config) args(args ...string) []string {
	ret := []string{"-p", c.port}
	if c.user != "" {
		ret = append(ret, "-u", c.user)
	}
	if c.client != "" {
		ret = append(ret, "-c", c.client)
	}
	return append(ret, args...)
}

func hasControlChar(s string) bool {
	return strings.IndexFunc(s, unicode.IsControl) >= 0
}

// p4ConfigFromURL returns the connection settings and the depot path of a
// p4:// or p4+ssl:// URL. Other schemes (for example https:// with
// "ghq.<url>.vcs = p4") are treated as plain TCP.
func p4ConfigFromURL(u *url.URL) (*p4Config, string, error) {
	if u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, "", fmt.Errorf("invalid p4 URL %q: queries, fragments and revision specifiers are not supported", u.Redacted())
	}
	if strings.HasPrefix(u.Scheme, "p4+") && u.Scheme != "p4+ssl" {
		return nil, "", fmt.Errorf("unsupported scheme %q: use p4:// or p4+ssl://", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return nil, "", fmt.Errorf("missing p4 server host in URL %q", u)
	}
	if hasControlChar(host) || strings.ContainsAny(host, "/@") {
		return nil, "", fmt.Errorf("invalid p4 server host in URL %q", u)
	}
	port := u.Port()
	if port == "" {
		port = p4DefaultPort
	}
	proto := ""
	if u.Scheme == "p4+ssl" {
		proto = "ssl"
	}
	if strings.Contains(host, ":") { // IPv6 literal
		if proto == "" {
			proto = "tcp"
		}
		proto += "6"
	}
	p4port := net.JoinHostPort(host, port)
	if proto != "" {
		p4port = proto + ":" + p4port
	}
	if _, _, err := p4ParsePort(p4port); err != nil {
		return nil, "", fmt.Errorf("invalid p4 server in URL %q: %w", u.Redacted(), err)
	}

	cfg := &p4Config{port: p4port}
	if u.User != nil {
		if _, ok := u.User.Password(); ok {
			return nil, "", fmt.Errorf("passwords in p4 URLs are not supported, use 'p4 login' instead: %q", u.Redacted())
		}
		cfg.user = u.User.Username()
		if hasControlChar(cfg.user) || strings.ContainsAny(cfg.user, " @#") {
			return nil, "", fmt.Errorf("invalid p4 user in URL %q", u.Redacted())
		}
	}

	depotPath, err := p4DepotPathFromURL(u)
	if err != nil {
		return nil, "", err
	}
	return cfg, depotPath, nil
}

// p4DepotPathFromURL converts the URL path "/depot/proj/main" to the depot
// path "//depot/proj/main", rejecting paths that p4 would interpret as
// wildcards or revision specifiers, or that could not be written into a
// client spec safely.
func p4DepotPathFromURL(u *url.URL) (string, error) {
	p := strings.TrimRight(u.Path, "/")
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return "", fmt.Errorf("missing depot path in URL %q", u.Redacted())
	}
	for _, part := range strings.Split(p, "/") {
		switch {
		case part == "", part == ".", part == "..":
			return "", fmt.Errorf("invalid depot path in URL %q: empty or dot path components are not allowed", u.Redacted())
		case strings.Contains(part, "..."), strings.ContainsAny(part, `*@#%"`):
			return "", fmt.Errorf("invalid depot path in URL %q: p4 wildcards and revision specifiers are not allowed", u.Redacted())
		case hasControlChar(part):
			return "", fmt.Errorf("invalid depot path in URL %q: control characters are not allowed", u.Redacted())
		}
	}
	return "//" + p, nil
}

var (
	p4NameUnsafeReg = regexp.MustCompile(`[^A-Za-z0-9_-]+`)
	p4Hostname      = os.Hostname // replaced in tests
)

func p4NamePart(s string, maxLen int) string {
	s = strings.Trim(p4NameUnsafeReg.ReplaceAllString(s, "_"), "_-")
	if len(s) > maxLen {
		s = strings.TrimRight(s[:maxLen], "_-")
	}
	return s
}

// p4ClientName returns a deterministic client name of the form
// <user>_<hostname>_<depot-path-slug>_<hash>. Sanitizing and truncating can
// map different inputs to the same readable parts, so the hash of all inputs
// keeps names unique.
func p4ClientName(user, hostname, port, depotPath, root string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{user, hostname, port, depotPath, root}, "\x00")))
	var parts []string
	for _, p := range []string{
		p4NamePart(user, 32), p4NamePart(hostname, 32), p4NamePart(depotPath, 64),
	} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "ghq")
	}
	return strings.Join(append(parts, hex.EncodeToString(sum[:4])), "_")
}

func p4LocalUser(u *url.URL) string {
	if u.User != nil && u.User.Username() != "" {
		return u.User.Username()
	}
	for _, env := range []string{"P4USER", "USER", "USERNAME"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}

// p4SpecField returns the value lines of a field in a p4 form such as the
// output of "p4 client -o". It returns nil if the field does not exist.
func p4SpecField(spec, name string) []string {
	var values []string
	found := false
	for _, line := range strings.Split(spec, "\n") {
		line = strings.TrimRight(line, "\r")
		if found {
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
				if v := strings.TrimSpace(line); v != "" {
					values = append(values, v)
				}
				continue
			}
			if line == "" {
				continue
			}
			break
		}
		if v, ok := strings.CutPrefix(line, name+":"); ok {
			found = true
			values = []string{}
			if v = strings.TrimSpace(v); v != "" {
				values = append(values, v)
			}
		}
	}
	if !found {
		return nil
	}
	return values
}

// p4SetSpecField replaces the value of a field in a p4 form, or appends the
// field if the form does not have it. With multiLine, values are written as
// indented lines below the field name.
func p4SetSpecField(spec, name string, multiLine bool, values ...string) string {
	field := []string{name + ":\t" + strings.Join(values, " ")}
	if multiLine {
		field = []string{name + ":"}
		for _, v := range values {
			field = append(field, "\t"+v)
		}
	}
	var out []string
	found, skipping := false, false
	for _, line := range strings.Split(spec, "\n") {
		line = strings.TrimRight(line, "\r")
		if skipping {
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
				continue
			}
			skipping = false
		}
		if strings.HasPrefix(line, name+":") && !found {
			found, skipping = true, true
			out = append(out, field...)
			continue
		}
		out = append(out, line)
	}
	if !found {
		out = append(out, "")
		out = append(out, field...)
	}
	return strings.Join(out, "\n")
}

// p4View maps depotPath into the workspace root, excluding a .p4config in the
// depot so that it can neither clobber nor be clobbered by ghq's own.
func p4View(depotPath, client string) []string {
	return []string{
		fmt.Sprintf(`"%s/..." "//%s/..."`, depotPath, client),
		fmt.Sprintf(`"-%s/%s" "//%s/%s"`, depotPath, p4ConfigName, client, p4ConfigName),
	}
}

// p4ViewDepotPath returns the depot path of the first inclusive mapping of
// a client view, e.g. "//depot/proj" for `//depot/proj/... //client/...`.
func p4ViewDepotPath(view []string) (string, error) {
	for _, line := range view {
		var lhs string
		if rest, ok := strings.CutPrefix(line, `"`); ok {
			lhs, _, _ = strings.Cut(rest, `"`)
		} else {
			lhs, _, _ = strings.Cut(line, " ")
		}
		if !strings.HasPrefix(lhs, "//") {
			continue // exclusion (-), overlay (+) or ditto (&) mapping
		}
		if p, ok := strings.CutSuffix(lhs, "/..."); ok && len(p) > 2 {
			return p, nil
		}
	}
	return "", fmt.Errorf("no depot path found in client view %q", view)
}

var p4HostReg = regexp.MustCompile(`^[A-Za-z0-9_]([A-Za-z0-9_.-]*[A-Za-z0-9_])?$`)

// p4ParsePort parses the P4PORT forms ghq supports, [tcp*:|ssl*:][host:]port,
// and returns the URL scheme and host:port. Other transports are rejected:
// with "rsh:", for instance, p4 runs a local command instead of connecting.
func p4ParsePort(port string) (scheme, hostport string, err error) {
	scheme, rest := "p4", port
	if proto, r, ok := strings.Cut(port, ":"); ok {
		switch proto {
		case "tcp", "tcp4", "tcp6", "tcp46", "tcp64":
			rest = r
		case "ssl", "ssl4", "ssl6", "ssl46", "ssl64":
			scheme, rest = "p4+ssl", r
		case "rsh", "jsh":
			return "", "", fmt.Errorf("unsupported P4PORT %q", port)
		}
	}
	host, num, splitErr := net.SplitHostPort(rest)
	if splitErr != nil {
		host, num = "localhost", rest // P4PORT=1666
	}
	if host == "" {
		host = "localhost"
	}
	if n, err := strconv.Atoi(num); err != nil || n < 1 || n > 65535 || strings.Trim(num, "0123456789") != "" ||
		(net.ParseIP(host) == nil && !p4HostReg.MatchString(host)) {
		return "", "", fmt.Errorf("unsupported P4PORT %q", port)
	}
	return scheme, net.JoinHostPort(host, num), nil
}

// p4SamePort reports whether two P4PORTs address the same server.
func p4SamePort(a, b string) bool {
	schemeA, hostA, errA := p4ParsePort(a)
	schemeB, hostB, errB := p4ParsePort(b)
	return errA == nil && errB == nil && schemeA == schemeB && strings.EqualFold(hostA, hostB)
}

// p4URLFromConfig rebuilds the p4:// URL of a workspace.
func p4URLFromConfig(cfg *p4Config, depotPath string) (*url.URL, error) {
	scheme, hostport, err := p4ParsePort(cfg.port)
	if err != nil {
		return nil, err
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   hostport,
		Path:   "/" + strings.TrimLeft(depotPath, "/"),
	}
	if cfg.user != "" {
		u.User = url.User(cfg.user)
	}
	return u, nil
}

func readP4Config(dir string) (*p4Config, error) {
	b, err := os.ReadFile(filepath.Join(dir, p4ConfigName))
	if err != nil {
		return nil, err
	}
	cfg := &p4Config{}
	for _, line := range strings.Split(string(b), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || strings.HasPrefix(key, "#") {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "P4PORT":
			cfg.port = value
		case "P4USER":
			cfg.user = value
		case "P4CLIENT":
			cfg.client = value
		}
	}
	name := filepath.Join(dir, p4ConfigName)
	if cfg.port == "" || cfg.client == "" {
		return nil, fmt.Errorf("%s does not set both P4PORT and P4CLIENT", name)
	}
	if _, _, err := p4ParsePort(cfg.port); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	for _, v := range []string{cfg.client, cfg.user} {
		if strings.HasPrefix(v, "-") || strings.IndexFunc(v, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
			return nil, fmt.Errorf("%s: invalid P4CLIENT or P4USER %q", name, v)
		}
	}
	return cfg, nil
}

// isP4Workspace reports whether dir has a .p4config naming a client. A
// .p4config that only sets shared settings such as P4PORT is not a workspace.
func isP4Workspace(dir string) bool {
	_, err := readP4Config(dir)
	return err == nil
}

func writeP4Config(dir string, cfg *p4Config) error {
	content := "P4PORT=" + cfg.port + "\n"
	if cfg.user != "" {
		content += "P4USER=" + cfg.user + "\n"
	}
	content += "P4CLIENT=" + cfg.client + "\n"
	return os.WriteFile(filepath.Join(dir, p4ConfigName), []byte(content), 0644)
}

func p4Error(err error, stderr *bytes.Buffer) error {
	if msg := strings.TrimSpace(stderr.String()); msg != "" {
		return fmt.Errorf("%w: %s", err, msg)
	}
	return err
}

// p4Output runs p4 silently and returns its standard output.
func p4Output(cfg *p4Config, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("p4", cfg.args(args...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmdutil.RunCommand(cmd, true); err != nil {
		return "", p4Error(err, &stderr)
	}
	return stdout.String(), nil
}

// p4SaveClient saves a client spec with "p4 client -i".
func p4SaveClient(cfg *p4Config, spec string, silent bool) error {
	var stderr bytes.Buffer
	cmd := exec.Command("p4", cfg.args("client", "-i")...)
	cmd.Stdin = strings.NewReader(spec)
	cmd.Stdout = os.Stderr
	if silent {
		cmd.Stdout = io.Discard
	}
	cmd.Stderr = &stderr
	if err := cmdutil.RunCommand(cmd, silent); err != nil {
		return p4Error(err, &stderr)
	}
	return nil
}

// p4ExistingClient returns the spec of cfg.client and whether the client
// already exists on the server. "p4 client -o" prints a default spec for a
// nonexistent client; only existing ones have an Access date.
func p4ExistingClient(cfg *p4Config) (string, bool, error) {
	spec, err := p4Output(cfg, "client", "-o", cfg.client)
	if err != nil {
		return "", false, err
	}
	return spec, p4SpecField(spec, "Access") != nil, nil
}

func p4SpecRoot(spec string) string {
	if v := p4SpecField(spec, "Root"); len(v) > 0 {
		return v[0]
	}
	return ""
}

func p4SameRoot(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	fa, errA := os.Stat(a)
	fb, errB := os.Stat(b)
	return errA == nil && errB == nil && os.SameFile(fa, fb)
}

// p4WorkspaceClient returns the spec of cfg.client after checking that the
// client exists and is rooted at root: "p4 sync" writes to the root recorded
// on the server, wherever it is run from.
func p4WorkspaceClient(cfg *p4Config, root string) (string, error) {
	spec, exists, err := p4ExistingClient(cfg)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("p4 client %q does not exist", cfg.client)
	}
	if r := p4SpecRoot(spec); !p4SameRoot(r, root) {
		return "", fmt.Errorf("the root of p4 client %q is %q, not %q", cfg.client, r, root)
	}
	return spec, nil
}

// p4RootMove repoints the client of a workspace moved by "ghq migrate".
type p4RootMove struct {
	cfg  *p4Config
	root string // as recorded on the server before the move
}

// p4PrepareRootMove checks, before dir is moved, that dir is the root of
// the client named in its .p4config.
func p4PrepareRootMove(dir string) (*p4RootMove, error) {
	cfg, err := readP4Config(dir)
	if err != nil {
		return nil, err
	}
	spec, err := p4WorkspaceClient(cfg, dir)
	if err != nil {
		return nil, err
	}
	return &p4RootMove{cfg: cfg, root: p4SpecRoot(spec)}, nil
}

// apply sets the root of the client to newRoot. Errors wrap
// errP4RootUnchanged when the server is known to keep the old root.
func (m *p4RootMove) apply(newRoot string) error {
	spec, exists, err := p4ExistingClient(m.cfg)
	if err != nil {
		// Nothing was sent, and the root was m.root when checked before the move.
		return fmt.Errorf("%w: %w", errP4RootUnchanged, err)
	}
	switch root := p4SpecRoot(spec); {
	case exists && root == newRoot:
		return nil
	case !exists || root != m.root:
		return fmt.Errorf("p4 client %q was changed or deleted meanwhile (root %q)", m.cfg.client, root)
	}
	saveErr := p4SaveClient(m.cfg, p4SetSpecField(spec, "Root", false, newRoot), true)
	if saveErr == nil {
		return nil
	}
	// p4 may fail after the server saved the spec, e.g. if its reply is lost.
	spec, exists, err = p4ExistingClient(m.cfg)
	switch {
	case err != nil:
		return fmt.Errorf("%w (checking the client afterwards failed: %v)", saveErr, err)
	case exists && p4SpecRoot(spec) == newRoot:
		return nil
	case exists && p4SpecRoot(spec) == m.root:
		return fmt.Errorf("%w: %w", errP4RootUnchanged, saveErr)
	}
	return saveErr
}

func p4Clone(vg *vcsGetOption) (err error) {
	if vg.branch != "" {
		return errP4Branch
	}
	if vg.bare {
		return errors.New("bare checkouts are unavailable for Perforce")
	}
	cfg, depotPath, err := p4ConfigFromURL(vg.url)
	if err != nil {
		return err
	}
	if hasControlChar(vg.dir) {
		return fmt.Errorf("invalid workspace root %q", vg.dir)
	}
	hostname, _ := p4Hostname()
	cfg.client = p4ClientName(p4LocalUser(vg.url), hostname, cfg.port, depotPath, vg.dir)

	// Claim the workspace directory: a concurrent "ghq get" of the same
	// path fails here. Until the sync starts, the directory holds at most
	// our .p4config, so a failure removes just that and the directory if it
	// is still empty, keeping e.g. a nested checkout made meanwhile.
	if err := os.MkdirAll(filepath.Dir(vg.dir), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(vg.dir, 0755); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists", vg.dir)
		}
		return err
	}
	syncing := false
	defer func() {
		if err != nil && !syncing {
			os.Remove(filepath.Join(vg.dir, p4ConfigName))
			os.Remove(vg.dir)
		}
	}()

	spec, exists, err := p4ExistingClient(cfg)
	if err != nil {
		return err
	}
	if exists {
		// A leftover client of the same workspace, e.g. after "ghq rm".
		// Anything else is not ours to modify.
		if root := p4SpecRoot(spec); !p4SameRoot(root, vg.dir) {
			return fmt.Errorf("p4 client %q already exists with root %q", cfg.client, root)
		}
	}
	spec = p4SetSpecField(spec, "Root", false, vg.dir)
	spec = p4SetSpecField(spec, "View", true, p4View(depotPath, cfg.client)...)
	if err := p4SaveClient(cfg, spec, vg.silent); err != nil {
		return err
	}
	if err := writeP4Config(vg.dir, cfg); err != nil {
		return err
	}
	if exists {
		// The have list of a reused client describes files that are gone.
		// Clear it (without touching files) rather than "sync -f", so that a
		// sync interrupted below can be resumed by a plain "p4 sync".
		if err := runInDir(vg.silent)(vg.dir, "p4", cfg.args("sync", "-k", "//"+cfg.client+"/...#none")...); err != nil {
			return err
		}
	}
	syncing = true
	if err := runInDir(vg.silent)(vg.dir, "p4", cfg.args("sync")...); err != nil {
		// Keep what was synced; the client and .p4config are in place.
		return fmt.Errorf("%w; run 'ghq get -u' to resume syncing %s", err, vg.dir)
	}
	if entries, err := os.ReadDir(vg.dir); err == nil && len(entries) == 1 {
		logger.Log("warning", fmt.Sprintf("no files found in %s/...; the workspace is empty", depotPath))
	}
	return nil
}

// PerforceBackend is the VCSBackend for Perforce (p4)
var PerforceBackend = &VCSBackend{
	// --shallow, --partial and --recursive are ignored: a workspace holds
	// only the synced revisions anyway.
	Clone: p4Clone,
	Update: func(vg *vcsGetOption) error {
		if vg.branch != "" {
			return errP4Branch
		}
		cfg, err := readP4Config(vg.dir)
		if err != nil {
			return err
		}
		// ghq paths omit the port, the user and a trailing ".git", so URLs
		// of other servers, users or depot paths can resolve to this
		// workspace. The server is named by p4:// URLs and explicit ports;
		// "ghq list | ghq get -u" passes https://<host>/<path>.
		var depotPath string
		if vg.url != nil {
			var want *p4Config
			want, depotPath, err = p4ConfigFromURL(vg.url)
			if err != nil {
				return err
			}
			namesPort := vg.url.Scheme == "p4" || vg.url.Scheme == "p4+ssl" || vg.url.Port() != ""
			if (namesPort && !p4SamePort(want.port, cfg.port)) || (want.user != "" && want.user != cfg.user) {
				return fmt.Errorf("%s is a workspace of P4PORT=%s P4USER=%s, not of %s", vg.dir, cfg.port, cfg.user, vg.url.Redacted())
			}
		}
		spec, err := p4WorkspaceClient(cfg, vg.dir)
		if err != nil {
			return err
		}
		if depotPath != "" {
			// The URL may also name a directory inside the workspace.
			if got, _ := p4ViewDepotPath(p4SpecField(spec, "View")); got == "" || (depotPath != got && !strings.HasPrefix(depotPath, got+"/")) {
				return fmt.Errorf("p4 client %q of %s maps %q, not %q", cfg.client, vg.dir, got, depotPath)
			}
		}
		return runInDir(vg.silent)(vg.dir, "p4", cfg.args("sync")...)
	},
	// Init is unsupported: depot paths are created by submitting files.
	Contents: []string{p4ConfigName},
	RemoteURL: func(dir string) (string, error) {
		cfg, err := readP4Config(dir)
		if err != nil {
			return "", err
		}
		spec, err := p4WorkspaceClient(cfg, dir)
		if err != nil {
			return "", err
		}
		depotPath, err := p4ViewDepotPath(p4SpecField(spec, "View"))
		if err != nil {
			return "", err
		}
		u, err := p4URLFromConfig(cfg, depotPath)
		if err != nil {
			return "", err
		}
		return u.String(), nil
	},
}
