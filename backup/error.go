package backup

import "github.com/giantswarm/microerror"

var invalidProviderError = microerror.New("invalid provider")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidProvider(err error) bool {
	return microerror.Cause(err) == invalidProviderError
}

var failedBackupError = microerror.New("backup failed")

// IsInvalidConfig asserts invalidConfigError.
func IsFailedBackupError(err error) bool {
	return microerror.Cause(err) == failedBackupError
}
