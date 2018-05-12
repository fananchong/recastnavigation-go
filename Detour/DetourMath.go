package detour

import "math"

func DtMathFabsf(x float64) float64             { return math.Abs(x) }
func DtMathSqrtf(x float64) float64             { return math.Sqrt(x) }
func DtMathFloorf(x float64) float64            { return math.Floor(x) }
func DtMathCeilf(x float64) float64             { return math.Ceil(x) }
func DtMathCosf(x float64) float64              { return math.Cos(x) }
func DtMathSinf(x float64) float64              { return math.Sin(x) }
func DtMathAtan2f(y float64, x float64) float64 { return math.Atan2(y, x) }
