package main

// compensateECByTemperature calculates the compensated conductivity value based on the given electrical conductivity (ec) and temperature.
//
// Parameters:
// - ec: the electrical conductivity value in Siemens per meter (S/m).
// - temperature: the temperature value in degrees Celsius.
//
// Returns:
// - float64: the compensated conductivity value in milliSiemens per centimeter (mS/cm).

func compensateECByTemperature(ec float64, temperature float64) float64 {
	if temperature == 0 {
		return ec
	}

	// taken from https://www.e-education.psu.edu/eme527/content/lectures/Lecture18/Lecture18.html
	// K = 0.0165 / celsius
	// K is in 1/ohm
	// conductivity in mS/cm
	k := 0.0165 / temperature
	compensatedConductivity := 1 / (1/(ec*1e-4) + k) * 1e4
	return compensatedConductivity
}
