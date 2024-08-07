// Code generated by "stringer -type=DataID"; DO NOT EDIT.

package store

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Humidity-0]
	_ = x[Temperature-1]
	_ = x[Conductivity-2]
	_ = x[Salinity-3]
	_ = x[TDS-4]
	_ = x[ConductivityWeighted-5]
	_ = x[ConductivityRaw-6]
}

const _DataID_name = "HumidityTemperatureConductivitySalinityTDSConductivityWeightedConductivityRaw"

var _DataID_index = [...]uint8{0, 8, 19, 31, 39, 42, 62, 77}

func (i DataID) String() string {
	if i < 0 || i >= DataID(len(_DataID_index)-1) {
		return "DataID(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DataID_name[_DataID_index[i]:_DataID_index[i+1]]
}
