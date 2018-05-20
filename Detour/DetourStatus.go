package detour

type DtStatus uint

const (

	// High level status.
	DT_FAILURE     DtStatus = 1 << 31 // Operation failed.
	DT_SUCCESS     DtStatus = 1 << 30 // Operation succeed.
	DT_IN_PROGRESS DtStatus = 1 << 29 // Operation still in progress.

	// Detail information for status.
	DT_STATUS_DETAIL_MASK DtStatus = 0x0ffffff
	DT_WRONG_MAGIC        DtStatus = 1 << 0 // Input data is not recognized.
	DT_WRONG_VERSION      DtStatus = 1 << 1 // Input data is in wrong version.
	DT_OUT_OF_MEMORY      DtStatus = 1 << 2 // Operation ran out of memory.
	DT_INVALID_PARAM      DtStatus = 1 << 3 // An input parameter was invalid.
	DT_BUFFER_TOO_SMALL   DtStatus = 1 << 4 // Result buffer for the query was too small to store all results.
	DT_OUT_OF_NODES       DtStatus = 1 << 5 // Query ran out of nodes during search.
	DT_PARTIAL_RESULT     DtStatus = 1 << 6 // Query did not reach the end location, returning best guess.
	DT_ALREADY_OCCUPIED   DtStatus = 1 << 7 // A tile has already been assigned to the given x,y coordinate
)

// Returns true of status is success.
func DtStatusSucceed(status DtStatus) bool {
	return (status & DT_SUCCESS) != 0
}

// Returns true of status is failure.
func DtStatusFailed(status DtStatus) bool {
	return (status & DT_FAILURE) != 0
}

// Returns true of status is in progress.
func DtStatusInProgress(status DtStatus) bool {
	return (status & DT_IN_PROGRESS) != 0
}

// Returns true if specific detail is set.
func DtStatusDetail(status DtStatus, detail DtStatus) bool {
	return (status & detail) != 0
}
