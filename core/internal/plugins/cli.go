package plugins

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// HostVersion is the launcher version the CLI checks minHost against. The launcher build sets it.
var HostVersion = "2.0.0"

const usageText = `Usage:
  plugin list
  plugin install <path-or-url> [--allow-unsigned]
  plugin update <id> [--allow-unsigned]
  plugin remove <id> [--delete-data]
  plugin enable <id>
  plugin disable <id>
  plugin rollback <id>
`

var errUsage = errors.New("usage")

// RunCLI runs one "plugin" command, such as "plugin list", and returns the exit code.
func RunCLI(args []string, out io.Writer) int {
	host, err := NewHost(HostOptions{Layout: DefaultLayout(), HostVersion: HostVersion})
	if err != nil {
		fmt.Fprintf(out, "error: %v\n", err)
		return 1
	}
	defer host.Shutdown()
	return runCommand(context.Background(), host, args, out)
}

func runCommand(ctx context.Context, host *Host, args []string, out io.Writer) int {
	if len(args) < 2 || args[0] != "plugin" {
		fmt.Fprint(out, usageText)
		return 2
	}
	err := dispatch(ctx, host, args[1], args[2:], out)
	switch {
	case errors.Is(err, errUsage):
		fmt.Fprint(out, usageText)
		return 2
	case err != nil:
		fmt.Fprintf(out, "error: %s\n", describeError(err))
		return 1
	default:
		return 0
	}
}

func dispatch(ctx context.Context, host *Host, command string, rest []string, out io.Writer) error {
	positional, flags, err := splitFlags(rest)
	if err != nil {
		return errUsage
	}
	if command == "list" {
		printList(host.List(), out)
		return nil
	}
	if len(positional) != 1 {
		return errUsage
	}
	target := positional[0]
	switch command {
	case "install":
		return cliInstall(ctx, host, target, flags["--allow-unsigned"], out)
	case "update":
		info, err := host.Update(ctx, UpdateRequest{ID: target, AllowUnsigned: flags["--allow-unsigned"]})
		return reportChange(info, err, "updated", out)
	case "rollback":
		info, err := host.Rollback(target)
		return reportChange(info, err, "rolled back", out)
	case "enable":
		info, err := host.SetEnabled(target, true)
		return reportChange(info, err, "enabled", out)
	case "disable":
		info, err := host.SetEnabled(target, false)
		return reportChange(info, err, "disabled", out)
	case "remove":
		if err := host.Remove(target, flags["--delete-data"]); err != nil {
			return err
		}
		fmt.Fprintf(out, "removed %s\n", target)
		return nil
	default:
		return errUsage
	}
}

var knownFlags = map[string]bool{"--allow-unsigned": true, "--delete-data": true}

// splitFlags separates "--flag" arguments from positional ones. Unknown flags are usage errors.
func splitFlags(args []string) ([]string, map[string]bool, error) {
	var positional []string
	flags := map[string]bool{}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}
		if !knownFlags[arg] {
			return nil, nil, errUsage
		}
		flags[arg] = true
	}
	return positional, flags, nil
}

// cliInstall accepts the permissions the package asks for. The command line is the consent step.
func cliInstall(ctx context.Context, host *Host, source string, allowUnsigned bool, out io.Writer) error {
	preview, err := host.PermissionsPreview(ctx, source, "")
	if err != nil {
		return err
	}
	info, err := host.Install(ctx, InstallRequest{Source: source, AllowUnsigned: allowUnsigned, AcceptedPermissions: preview.Consent})
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "installed %s %s\n", info.ID, info.Version)
	printPermissions(info.Permissions, out)
	if info.Unsigned {
		fmt.Fprintln(out, "warning: the package is not signed")
	}
	return nil
}

func reportChange(info PluginInfo, err error, verb string, out io.Writer) error {
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s %s (version %s)\n", verb, info.ID, info.Version)
	return nil
}

func printList(infos []PluginInfo, out io.Writer) {
	if len(infos) == 0 {
		fmt.Fprintln(out, "no plugins installed")
		return
	}
	fmt.Fprintf(out, "%-16s %-10s %s\n", "ID", "VERSION", "STATUS")
	for _, info := range infos {
		fmt.Fprintf(out, "%-16s %-10s %s\n", info.ID, info.Version, info.Status)
	}
}

func printPermissions(p Permissions, out io.Writer) {
	fmt.Fprintf(out, "network: %s\n", listOrNone(p.Network))
	fmt.Fprintf(out, "folders: %s\n", listOrNone(p.FS))
	fmt.Fprintf(out, "secrets: %t, browser: %t, exec: %t\n", p.Secrets, p.Browser, p.Exec)
}

func listOrNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, ", ")
}

// describeError turns the sentinel errors into a sentence the user can act on.
func describeError(err error) string {
	if errors.Is(err, ErrUnsigned) {
		return "the package is not signed by a trusted key. Add --allow-unsigned to install it anyway"
	}
	return err.Error()
}
