package version_checker

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
	"regexp"
	"strconv"
	"strings"
)

type VersionCheckerBinary int

// List of binaries supported for version checking.
const (
	Docker VersionCheckerBinary = iota
	Terraform
	Packer
)

const (
	// VersionRegexMatcher is a regex used to extract version string from shell command output.
	VersionRegexMatcher = "\\d+(\\.\\d+)+"
	// DefaultVersionArg is a default arg to pass in to get version output from shell command.
	DefaultVersionArg = "--version"
)

type CheckVersionParams struct {
	binary          VersionCheckerBinary
	expectedVersion string
	workingDir      string
}

// CheckVersionE checks whether the given binary version is greater or equal
// to the given expected version.
func CheckVersionE(
	t testing.TestingT,
	params CheckVersionParams) error {

	err := validateParams(params)
	if err != nil {
		return err
	}

	version, err := getVersionWithShellCommand(t, params)
	if err != nil {
		return err
	}

	return validateBinaryVersionGreaterOrEqual(version, params.expectedVersion)
}

// CheckVersion checks whether the given binary version is greater or equal to the
// given expected version and fail if it's not.
func CheckVersion(
	t testing.TestingT,
	params CheckVersionParams) {
	err := CheckVersionE(t, params)
	require.NoError(t, err)
}

// Validate whether the given params contains valid data to check version.
func validateParams(params CheckVersionParams) error {
	// Check for empty parameters
	if len(params.expectedVersion) == 0 {
		return fmt.Errorf("empty expectedVersion found")
	} else if len(params.workingDir) == 0 {
		return fmt.Errorf("empty workingDir found")
	} else if params.binary < 0 {
		return fmt.Errorf("empty binary found")
	}

	// Check the format of the expected version.
	match, err := regexp.MatchString(VersionRegexMatcher, params.expectedVersion)
	if err != nil {
		return fmt.Errorf(
			"unexpected regex matcher compilation failure: %w", err)
	} else if !match {
		return fmt.Errorf(
			"invalid expected version format found {%s}", params.expectedVersion)
	}

	return nil
}

// getVersionWithShellCommand get version by running a shell command.
func getVersionWithShellCommand(t testing.TestingT, params CheckVersionParams) (string, error) {
	// Set appropriate binary, versionArg and extract version function
	// based on the given parameters.
	var binaryName, versionArg = "", DefaultVersionArg
	switch params.binary {
	case Docker:
		binaryName = "docker"
	case Packer:
		binaryName = "packer"
	case Terraform:
		binaryName = "terraform"
	default:
		t.Fatalf("unsupported binary for checking versions.")
	}

	// Run a shell command to get the version string.
	output, err := shell.RunCommandAndGetOutputE(t, shell.Command{
		Command:    binaryName,
		Args:       []string{versionArg},
		WorkingDir: params.workingDir,
		Env:        map[string]string{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to run shell command for binaray {%s} "+
			"w/ version args {%s}: %w", binaryName, versionArg, err)
	}

	version, err := extractVersionFromShellCommandOutput(output)
	if err != nil {
		return "", fmt.Errorf("failed to extract version from shell "+
			"command output {%s}: %w", output, err)
	}

	return version, nil
}

// extractVersionFromShellCommandOutput extracts version with regex string matching
// from the given shell command output string.
func extractVersionFromShellCommandOutput(output string) (string, error) {
	regexMatcher, err := regexp.Compile(VersionRegexMatcher)
	if err != nil {
		return "", fmt.Errorf("unexpected regex compilation error: %w", err)
	}

	version := regexMatcher.FindString(output)
	if len(version) == 0 {
		return "", fmt.Errorf("failed to find version using regex matcher")
	}

	return version, nil
}

// validateBinaryVersionGreaterOrEqual asserts that the first version is greater
// than or equal to the second version.
//
//    validateBinaryVersionGreaterOrEqual(t, 1.0.31, 1.0.27) - no error
//    validateBinaryVersionGreaterOrEqual(t, 1.0.10, 1.0.27) - error
//    validateBinaryVersionGreaterOrEqual(t, 1.0, 1.0.10) - error
func validateBinaryVersionGreaterOrEqual(versionStr1 string, versionStr2 string) error {
	if len(versionStr1) == 0 {
		return fmt.Errorf("empty versionStr1 found")
	} else if len(versionStr2) == 0 {
		return fmt.Errorf("empty versionStr2 found")
	}

	var versionNums1 = strings.Split(versionStr1, ".")
	var versionNums2 = strings.Split(versionStr2, ".")
	v1Length, v2Length := len(versionNums1), len(versionNums2)

	length := v1Length
	if v2Length < v1Length {
		length = v2Length
	}

	for i := 0; i < length; i++ {
		v1, err := strconv.Atoi(versionNums1[i])
		if err != nil {
			return fmt.Errorf("invalid format version found {%s}", versionStr1)
		}

		v2, err := strconv.Atoi(versionNums2[i])
		if err != nil {
			return fmt.Errorf("invalid format version found {%s}", versionStr2)
		}

		if v1 < v2 {
			return fmt.Errorf("binary version {%s} is smaller than {%s}",
				versionStr1, versionStr2)
		} else if v1 > v2 {
			// version1 is higher than version2
			return nil
		}
	}

	// Final check to make sure the length of the version1 is
	// greater or equal than version2. (e.g., 1.0.12 vs. 1.0)
	if v1Length < v2Length {
		return fmt.Errorf("binary version {%s} is smaller than {%s}",
			versionStr1, versionStr2)
	}

	return nil
}
