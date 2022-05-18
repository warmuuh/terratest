package version_checker

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
	"regexp"
)

// VersionMismatchErr is an error to indicate version mismatch.
type VersionMismatchErr struct {
	expectedVersion string
	actualVersion   string
}

func (r *VersionMismatchErr) Error() string {
	return fmt.Sprintf("found version mismatch {%s} when "+
		"expecting min version of {%s}", r.expectedVersion, r.actualVersion)
}

// VersionCheckerBinary is an enum for supported version checking.
type VersionCheckerBinary int

// List of binaries supported for version checking.
const (
	Docker VersionCheckerBinary = iota
	Terraform
	Packer
)

const (
	// versionRegexMatcher is a regex used to extract version string from shell command output.
	versionRegexMatcher = `\d+(\.\d+)+`
	// defaultVersionArg is a default arg to pass in to get version output from shell command.
	defaultVersionArg = "--version"
)

type CheckVersionParams struct {
	Binary          VersionCheckerBinary
	ExpectedVersion string
	WorkingDir      string
}

// CheckVersionE checks whether the given Binary version is greater than or equal
// to the given expected version.
func CheckVersionE(
	t testing.TestingT,
	params CheckVersionParams) error {

	if err := validateParams(params); err != nil {
		return err
	}

	binaryVersion, err := getVersionWithShellCommand(t, params)
	if err != nil {
		return err
	}

	if err := checkMinimumBinaryVersion(binaryVersion, params.ExpectedVersion); err != nil {
		return err
	}

	return nil
}

// CheckVersion checks whether the given Binary version is greater than or equal to the
// given expected version and fail if it's not.
func CheckVersion(
	t testing.TestingT,
	params CheckVersionParams) {
	require.NoError(t, CheckVersionE(t, params))
}

// Validate whether the given params contains valid data to check version.
func validateParams(params CheckVersionParams) error {
	// Check for empty parameters
	if params.ExpectedVersion == "" {
		return fmt.Errorf("set ExpectedVersion in params")
	} else if params.WorkingDir == "" {
		return fmt.Errorf("set WorkingDir in params")
	} else if params.Binary < 0 {
		return fmt.Errorf("set Binary in params")
	}

	// Check the format of the expected version.
	if _, err := version.NewVersion(params.ExpectedVersion); err != nil {
		return fmt.Errorf(
			"invalid version format found {%s}", params.ExpectedVersion)
	}

	return nil
}

// getVersionWithShellCommand get version by running a shell command.
func getVersionWithShellCommand(t testing.TestingT, params CheckVersionParams) (string, error) {
	// Set appropriate Binary, versionArg and extract version function
	// based on the given parameters.
	var binaryName = ""
	var versionArg = defaultVersionArg
	switch params.Binary {
	case Docker:
		binaryName = "docker"
	case Packer:
		binaryName = "packer"
	case Terraform:
		binaryName = "terraform"
	default:
		t.Fatalf("unsupported Binary for checking versions {%s}.", params.Binary)
	}

	// Run a shell command to get the version string.
	output, err := shell.RunCommandAndGetOutputE(t, shell.Command{
		Command:    binaryName,
		Args:       []string{versionArg},
		WorkingDir: params.WorkingDir,
		Env:        map[string]string{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to run shell command for Binary {%s} "+
			"w/ version args {%s}: %w", binaryName, versionArg, err)
	}

	versionStr, err := extractVersionFromShellCommandOutput(output)
	if err != nil {
		return "", fmt.Errorf("failed to extract version from shell "+
			"command output {%s}: %w", output, err)
	}

	return versionStr, nil
}

// extractVersionFromShellCommandOutput extracts version with regex string matching
// from the given shell command output string.
func extractVersionFromShellCommandOutput(output string) (string, error) {
	regexMatcher := regexp.MustCompile(versionRegexMatcher)
	versionStr := regexMatcher.FindString(output)
	if versionStr == "" {
		return "", fmt.Errorf("failed to find version using regex matcher")
	}

	return versionStr, nil
}

// checkMinimumBinaryVersion checks whether the given version is greater
// than or equal to the given minimum version.
//
// It returns Error for ill-formatted version string and VersionMismatchErr for
// minimum version check failure.
//
//    checkMinimumBinaryVersion(t, 1.0.31, 1.0.27) - no error
//    checkMinimumBinaryVersion(t, 1.0.10, 1.0.27) - error
//    checkMinimumBinaryVersion(t, 1.0, 1.0.10) - error
func checkMinimumBinaryVersion(actualVersionStr string, minimumVersionStr string) error {
	version1, err := version.NewVersion(actualVersionStr)
	if err != nil {
		return fmt.Errorf("invalid version format found for actualVersionStr: %s", actualVersionStr)
	}

	version2, err := version.NewVersion(minimumVersionStr)
	if err != nil {
		return fmt.Errorf("invalid version format found for minimumVersionStr: %s", minimumVersionStr)
	}

	if version1.LessThan(version2) {
		return &VersionMismatchErr{
			expectedVersion: minimumVersionStr,
			actualVersion:   actualVersionStr,
		}
	}

	return nil
}
