package version_checker

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParamValidation(t *testing.T) {
	t.Parallel()

	// empty params
	err := validateParams(CheckVersionParams{})
	require.EqualError(t, err, "empty expectedVersion found")

	// invalid binary
	err = validateParams(CheckVersionParams{
		binary:          -1,
		expectedVersion: "1.2",
		workingDir:      ".",
	})
	require.EqualError(t, err, "empty binary found")

	// invalid working Dir
	err = validateParams(CheckVersionParams{
		binary:          Docker,
		expectedVersion: "1.2",
		workingDir:      "",
	})
	require.EqualError(t, err, "empty workingDir found")

	// invalid expected version
	err = validateParams(CheckVersionParams{
		binary:          Docker,
		expectedVersion: "abc",
		workingDir:      ".",
	})
	require.EqualError(t, err, "invalid expected version format found {abc}")

	// valid params
	err = validateParams(CheckVersionParams{
		binary:          Docker,
		expectedVersion: "1.2.3",
		workingDir:      ".",
	})
	require.NoError(t, err)
}

func TestExtractVersionFromShellCommandOutput(t *testing.T) {
	t.Parallel()

	// empty output
	_, err := extractVersionFromShellCommandOutput("")
	require.EqualError(t, err, "failed to find version using regex matcher")

	// invalid version output from shell command
	_, err = extractVersionFromShellCommandOutput("invalid output")
	require.EqualError(t, err, "failed to find version using regex matcher")
}

func TestValidateBinaryVersionGreaterOrEqual(t *testing.T) {
	t.Parallel()

	// empty versionStr1
	err := validateBinaryVersionGreaterOrEqual("", "")
	require.EqualError(t, err, "empty versionStr1 found")

	// empty versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.3", "")
	require.EqualError(t, err, "empty versionStr2 found")

	// invalid format versionStr1
	err = validateBinaryVersionGreaterOrEqual("abc", "1.2.3")
	require.EqualError(t, err, "invalid format version found {abc}")

	// invalid format versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.3", "abc")
	require.EqualError(t, err, "invalid format version found {abc}")

	// versionStr1 > versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.3", "1.2.2")
	require.NoError(t, err)

	// versionStr1 = versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.3", "1.2.3")
	require.NoError(t, err)

	// versionStr1 < versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.3", "1.2.4")
	require.EqualError(t, err, "binary version {1.2.3} is smaller than {1.2.4}")

	// mismatching version length: versionStr1 > versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2", "1.2.4")
	require.EqualError(t, err, "binary version {1.2} is smaller than {1.2.4}")

	// mismatching version length: versionStr1 < versionStr2
	err = validateBinaryVersionGreaterOrEqual("1.2.4", "1.2")
	require.NoError(t, err)
}

func TestCheckVersionSanityCheck(t *testing.T) {
	t.Parallel()

	// Note: with the current implementation of running shell command, it's not easy to
	// mock the output of running a shell command. So we assume a certain binary is installed in the working
	// directory and it's greater than 0.
	err := CheckVersionE(t, CheckVersionParams{
		binary:          Terraform,
		expectedVersion: "0.0.1",
		workingDir:      ".",
	})
	require.NoError(t, err)
}
