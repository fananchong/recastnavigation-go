package detour

/**
@defgroup detour Detour

Members in this module are used to create, manipulate, and query navigation
meshes.

@note This is a summary list of members.  Use the index or search
feature to find minor members.
*/

/// @name General helper functions
/// @{

/// Used to ignore a function parameter.  VS complains about unused parameters
/// and this silences the warning.
///  @param [in] _ Unused parameter
func DtIgnoreUnused(interface{}) {}

/// Swaps the values of the two parameters.
///  @param[in,out]	a	Value A
///  @param[in,out]	b	Value B
func DtSwapFloat64(a, b *float64) { t := *a; *a = *b; *b = t }
func DtSwapUInt32(a, b *uint32)   { t := *a; *a = *b; *b = t }
func DtSwapInt32(a, b *int32)     { t := *a; *a = *b; *b = t }
func DtSwapUInt16(a, b *uint16)   { t := *a; *a = *b; *b = t }
func DtSwapInt16(a, b *int16)     { t := *a; *a = *b; *b = t }

/// Returns the minimum of two values.
///  @param[in]		a	Value A
///  @param[in]		b	Value B
///  @return The minimum of the two values.
func DtMinFloat64(a, b float64) float64 {
	if a < b {
		return a
	} else {
		return b
	}
}
func DtMinUInt32(a, b uint32) uint32 {
	if a < b {
		return a
	} else {
		return b
	}
}
func DtMinInt32(a, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}
func DtMinUInt16(a, b uint16) uint16 {
	if a < b {
		return a
	} else {
		return b
	}
}
func DtMinInt16(a, b int16) int16 {
	if a < b {
		return a
	} else {
		return b
	}
}

/// Returns the maximum of two values.
///  @param[in]		a	Value A
///  @param[in]		b	Value B
///  @return The maximum of the two values.
func DtMaxFloat64(a, b float64) float64 {
	if a > b {
		return a
	} else {
		return b
	}
}
func DtMaxUInt32(a, b uint32) uint32 {
	if a > b {
		return a
	} else {
		return b
	}
}
func DtMaxInt32(a, b int32) int32 {
	if a > b {
		return a
	} else {
		return b
	}
}
func DtMaxUInt16(a, b uint16) uint16 {
	if a > b {
		return a
	} else {
		return b
	}
}
func DtMaxInt16(a, b int16) int16 {
	if a > b {
		return a
	} else {
		return b
	}
}

/// Returns the absolute value.
///  @param[in]		a	The value.
///  @return The absolute value of the specified value.
func DtAbsFloat64(a float64) float64 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}
func DtAbsInt32(a int32) int32 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}
func DtAbsInt16(a int16) int16 {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

/// Returns the square of the value.
///  @param[in]		a	The value.
///  @return The square of the value.
func DtSqrFloat64(a float64) float64 { return a * a }
func DtSqrUInt32(a uint32) uint32    { return a * a }
func DtSqrInt32(a int32) int32       { return a * a }
func DtSqrUInt16(a uint16) uint16    { return a * a }
func DtSqrInt16(a int16) int16       { return a * a }

/// Clamps the value to the specified range.
///  @param[in]		v	The value to clamp.
///  @param[in]		mn	The minimum permitted return value.
///  @param[in]		mx	The maximum permitted return value.
///  @return The value, clamped to the specified range.
func DtClampFloat64(v, mn, mx float64) float64 {
	if v < mn {
		return mn
	} else {
		if v > mx {
			return mx
		} else {
			return v
		}
	}
}
func DtClampUInt32(v, mn, mx uint32) uint32 {
	if v < mn {
		return mn
	} else {
		if v > mx {
			return mx
		} else {
			return v
		}
	}
}
func DtClampInt32(v, mn, mx int32) int32 {
	if v < mn {
		return mn
	} else {
		if v > mx {
			return mx
		} else {
			return v
		}
	}
}
func DtClampUInt16(v, mn, mx uint16) uint16 {
	if v < mn {
		return mn
	} else {
		if v > mx {
			return mx
		} else {
			return v
		}
	}
}
func DtClampInt16(v, mn, mx int16) int16 {
	if v < mn {
		return mn
	} else {
		if v > mx {
			return mx
		} else {
			return v
		}
	}
}

/// @}
/// @name Vector helper functions.
/// @{

/// Derives the cross product of two vectors. (@p v1 x @p v2)
///  @param[out]	dest	The cross product. [(x, y, z)]
///  @param[in]		v1		A Vector [(x, y, z)]
///  @param[in]		v2		A vector [(x, y, z)]
func DtVcross(dest, v1, v2 *[3]float64) {
	dest[0] = v1[1]*v2[2] - v1[2]*v2[1]
	dest[1] = v1[2]*v2[0] - v1[0]*v2[2]
	dest[2] = v1[0]*v2[1] - v1[1]*v2[0]
}
