package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// preparedPackage is an unpacked and verified package that is not installed yet.
type preparedPackage struct {
	Work     string // scratch folder that holds the download and the unpacked files
	Dir      string // root of the unpacked package
	Manifest Manifest
	Unsigned bool
}

func (p preparedPackage) discard() {
	_ = os.RemoveAll(p.Work)
}

// Preview is what the consent screen shows before an install.
type Preview struct {
	Manifest Manifest `json:"manifest"`
	Unsigned bool     `json:"unsigned"`
}

// preparePackage downloads (for URLs), unpacks and verifies a package without installing it.
// The caller must call discard on the result.
func (in *installer) preparePackage(ctx context.Context, source string) (prepared preparedPackage, err error) {
	if err := os.MkdirAll(in.layout.Root, 0o755); err != nil {
		return preparedPackage{}, err
	}
	work, err := os.MkdirTemp(in.layout.Root, ".staging-*")
	if err != nil {
		return preparedPackage{}, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(work)
		}
	}()

	archive, err := in.archiveFor(ctx, source, work)
	if err != nil {
		return preparedPackage{}, err
	}
	prepared = preparedPackage{Work: work, Dir: filepath.Join(work, "package")}
	if err := extractZip(archive, prepared.Dir); err != nil {
		return preparedPackage{}, err
	}
	manifest, checksums, err := verifyPackage(prepared.Dir)
	if err != nil {
		return preparedPackage{}, err
	}
	keys, err := loadTrustedKeys(in.layout.TrustedKeysPath())
	if err != nil {
		return preparedPackage{}, err
	}
	signed, err := verifySignature(prepared.Dir, checksums, keys)
	if err != nil {
		return preparedPackage{}, err
	}
	if err := checkCompatibility(manifest, prepared.Dir, in.hostVersion); err != nil {
		return preparedPackage{}, err
	}
	prepared.Manifest = manifest
	prepared.Unsigned = !signed
	return prepared, nil
}

// archiveFor returns a local path to the .llplugin file, downloading it when source is a URL.
func (in *installer) archiveFor(ctx context.Context, source, work string) (string, error) {
	if isURL(source) {
		return fetchPackage(ctx, in.client, source, work)
	}
	if _, err := os.Stat(source); err != nil {
		return "", fmt.Errorf("package not found: %w", err)
	}
	return source, nil
}

// verifyPackage reads plugin.json and checks every file against SHA256SUMS.
func verifyPackage(dir string) (Manifest, []byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if err != nil {
		return Manifest{}, nil, fmt.Errorf("package has no plugin.json")
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return Manifest{}, nil, err
	}
	checksums, err := verifyChecksums(dir)
	if err != nil {
		return Manifest{}, nil, err
	}
	return manifest, checksums, nil
}

// checkCompatibility rejects packages this launcher cannot run and entries that are missing.
func checkCompatibility(manifest Manifest, dir, hostVersion string) error {
	if manifest.APIVersion != SupportedAPIVersion {
		return fmt.Errorf("plugin needs protocol %s, this launcher speaks %s", manifest.APIVersion, SupportedAPIVersion)
	}
	order, err := compareSemver(hostVersion, manifest.MinHost)
	if err != nil {
		return fmt.Errorf("launcher version %q: %w", hostVersion, err)
	}
	if order < 0 {
		return fmt.Errorf("plugin needs launcher %s or newer, this is %s", manifest.MinHost, hostVersion)
	}
	if _, ok := manifest.EntryFor(PlatformKey()); !ok {
		return fmt.Errorf("package has no build for %s", PlatformKey())
	}
	for key, entry := range manifest.Entry {
		if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(entry))); err != nil || info.IsDir() {
			return fmt.Errorf("entry %s points at %q, which is not in the package", key, entry)
		}
	}
	return nil
}
