package version_checker

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParamValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		param                CheckVersionParams
		containError         bool
		expectedErrorMessage string
	}{
		{
			param:                CheckVersionParams{},
			containError:         true,
			expectedErrorMessage: "set ExpectedVersion in params",
		},
		{
			param: CheckVersionParams{
				Binary:          -1,
				ExpectedVersion: "1.2",
				WorkingDir:      ".",
			},
			containError:         true,
			expectedErrorMessage: "set Binary in params",
		},
		{
			param: CheckVersionParams{
				Binary:          Docker,
				ExpectedVersion: "1.2",
				WorkingDir:      "",
			},
			containError:         true,
			expectedErrorMessage: "set WorkingDir in params",
		},
		{
			param: CheckVersionParams{
				Binary:          Docker,
				ExpectedVersion: "abc",
				WorkingDir:      ".",
			},
			containError:         true,
			expectedErrorMessage: "invalid version format found {abc}",
		},
		{
			param: CheckVersionParams{
				Binary:          Docker,
				ExpectedVersion: "1.2.3",
				WorkingDir:      ".",
			},
			containError:         false,
			expectedErrorMessage: "",
		},
	}

	for _, tc := range tests {
		err := validateParams(tc.param)
		testCaseContextStr := fmt.Sprintf("test case w/ param: %v", tc.param)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, testCaseContextStr)
		} else {
			require.NoError(t, err, testCaseContextStr)
		}
	}
}

func TestExtractVersionFromShellCommandOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		outputStr            string
		expectedVersionStr   string
		containError         bool
		expectedErrorMessage string
	}{
		{outputStr: "version is 1.2.3", expectedVersionStr: "1.2.3", containError: false,
			expectedErrorMessage: ""},
		{outputStr: "version is v1.0.0", expectedVersionStr: "1.0.0", containError: false,
			expectedErrorMessage: ""},
		{outputStr: "version is v1.0", expectedVersionStr: "1.0", containError: false,
			expectedErrorMessage: ""},
		{outputStr: "version is vabc", expectedVersionStr: "", containError: true,
			expectedErrorMessage: "failed to find version using regex matcher"},
		{outputStr: "", expectedVersionStr: "", containError: true,
			expectedErrorMessage: "failed to find version using regex matcher"},
	}

	for _, tc := range tests {
		versionStr, err := extractVersionFromShellCommandOutput(tc.outputStr)
		testCaseContextStr := fmt.Sprintf("test case w/ outputStr: %s", tc.outputStr)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, testCaseContextStr)
		} else {
			require.NoError(t, err, testCaseContextStr)
			require.Equal(t, tc.expectedVersionStr, versionStr, testCaseContextStr)
		}
	}
}

func TestValidateBinaryVersionGreaterOrEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		actualVersionStr     string
		minimumVersionStr    string
		containError         bool
		expectedErrorMessage string
	}{
		{actualVersionStr: "", minimumVersionStr: "1.2.3", containError: true,
			expectedErrorMessage: "invalid version format found for actualVersionStr: "},
		{actualVersionStr: "1.2.3", minimumVersionStr: "", containError: true,
			expectedErrorMessage: "invalid version format found for minimumVersionStr: "},
		{actualVersionStr: "1.2.3", minimumVersionStr: "1.2.3", containError: false,
			expectedErrorMessage: ""},
		{actualVersionStr: "1.2.4", minimumVersionStr: "1.2.3", containError: false,
			expectedErrorMessage: ""},
		{actualVersionStr: "1.2", minimumVersionStr: "1.2.3", containError: true,
			expectedErrorMessage: "found version mismatch {1.2.3} when expecting min version of {1.2}"},
	}

	for _, tc := range tests {
		err := checkMinimumBinaryVersion(tc.actualVersionStr, tc.minimumVersionStr)
		testCaseContextStr := fmt.Sprintf("test case w/ actualVersionStr: %s, expectedVersionstr: %s",
			tc.actualVersionStr, tc.minimumVersionStr)
		if tc.containError {
			require.EqualError(t, err, tc.expectedErrorMessage, testCaseContextStr)
		} else {
			require.NoError(t, err, testCaseContextStr)
		}
	}
}

func TestCheckVersionSanityCheck(t *testing.T) {
	t.Parallel()

	// Note: with the current implementation of running shell command, it's not easy to
	// mock the output of running a shell command. So we assume a certain Binary is installed in the working
	// directory and it's greater than 0.
	err := CheckVersionE(t, CheckVersionParams{
		Binary:          Terraform,
		ExpectedVersion: "0.0.1",
		WorkingDir:      ".",
	})
	require.NoError(t, err)
}
