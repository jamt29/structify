package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/jamt29/structify/internal/buildinfo"
	"github.com/spf13/cobra"
)

const (
	latestReleaseURL = "https://api.github.com/repos/jamt29/structify/releases/latest"
	updateInstallRef = "github.com/jamt29/structify/cmd/structify@latest"
	releasesPageURL  = "https://github.com/jamt29/structify/releases/latest"
)

var (
	updateCheck bool
	updateYes   bool

	fetchLatestReleaseFn = fetchLatestRelease
	runSelfUpdateFn      = runSelfUpdate
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

var errGoNotFound = errors.New("go binary not found in PATH")

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualizar Structify a la ultima version disponible",
	Long: "Verifica la ultima release publicada y permite actualizar el CLI sin salir de la terminal.\n\n" +
		"Ejemplos:\n" +
		"  structify update --check\n" +
		"  structify update\n" +
		"  structify update --yes",
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "solo verificar si hay una nueva version disponible")
	updateCmd.Flags().BoolVar(&updateYes, "yes", false, "actualizar sin pedir confirmacion")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	current := normalizeVersion(buildinfo.Version)
	release, err := fetchLatestReleaseFn()
	if err != nil {
		return err
	}
	latest := normalizeVersion(release.TagName)
	if latest == "" {
		return fmt.Errorf("latest release does not include tag_name")
	}

	fmt.Fprintf(out, "Version actual:  %s\n", displayVersion(current))
	fmt.Fprintf(out, "Ultima version:  %s\n\n", displayVersion(latest))

	if current != "" && !isNewer(current, latest) {
		fmt.Fprintf(out, "✓ Ya tienes la ultima version (%s)\n", displayVersion(latest))
		return nil
	}

	if updateCheck {
		fmt.Fprintln(out, "Hay una nueva version disponible.")
		fmt.Fprintln(out, "Ejecuta: structify update")
		return nil
	}

	if !updateYes {
		ok, err := confirmUpdate(cmd.InOrStdin(), out, latest)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "Actualizacion cancelada.")
			return nil
		}
	}

	fmt.Fprintln(out, "\nActualizando...")
	if err := runSelfUpdateFn(); err != nil {
		if errors.Is(err, errGoNotFound) {
			fmt.Fprintln(out, "No se encontro 'go' en el PATH.")
			fmt.Fprintf(out, "Descarga el binario desde:\n%s\n", releasesPageURL)
			return nil
		}
		return err
	}

	fmt.Fprintf(out, "✓ Structify actualizado a %s\n\n", displayVersion(latest))
	fmt.Fprintln(out, "Reinicia la terminal para usar la nueva version.")
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating update request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting latest release: unexpected status %s", resp.Status)
	}

	var payload githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decoding latest release response: %w", err)
	}
	return &payload, nil
}

func runSelfUpdate() error {
	if _, err := exec.LookPath("go"); err != nil {
		return errGoNotFound
	}
	c := exec.Command("go", "install", updateInstallRef)
	output, err := c.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			return fmt.Errorf("updating structify: %w", err)
		}
		return fmt.Errorf("updating structify: %s", msg)
	}
	return nil
}

func confirmUpdate(in io.Reader, out io.Writer, latest string) (bool, error) {
	fmt.Fprintf(out, "Actualizar a %s? [s/N]: ", displayVersion(latest))
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}
	v := strings.ToLower(strings.TrimSpace(line))
	return v == "s" || v == "si" || v == "sí" || v == "y" || v == "yes", nil
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "dev" || v == "unknown" {
		return ""
	}
	return v
}

func displayVersion(v string) string {
	if strings.TrimSpace(v) == "" {
		return "desconocida"
	}
	return "v" + v
}

// compareVersions compares semver-like versions (major.minor.patch).
// Returns -1 if a<b, 0 if a==b, 1 if a>b.
func compareVersions(a, b string) int {
	pa := parseVersionParts(a)
	pb := parseVersionParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func isNewer(current, latest string) bool {
	return compareVersions(current, latest) < 0
}

func parseVersionParts(v string) [3]int {
	var out [3]int
	parts := strings.Split(strings.TrimSpace(v), ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, r := range parts[i] {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		out[i] = n
	}
	return out
}

