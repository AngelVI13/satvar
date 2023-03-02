package gps

// calculateElevation Calculates elevation gain & loss give a slice of elevation data.
// Here chunkSize indicates the smoothing filter size. Since GPS tracks usually
// record at approx. once per second if you set a chunkSize of 60 -> each 60
// elevation points will be grouped together and one average value will be produced
// for them.
func calculateElevation(points []float64, chunkSize int) (gain, loss float64) {
	var temp = make([]float64, chunkSize)
	var avg []float64

	for i := 0; i < len(points); i++ {
		temp[i%chunkSize] = points[i]

		if i%chunkSize == 0 {
			avgElev := 0.0
			for _, elev := range temp {
				avgElev += elev
			}
			avgElev /= float64(chunkSize)
			avg = append(avg, avgElev)
		}
	}

	for i := 1; i < len(avg); i++ {
		diff := avg[i] - avg[i-1]

		if diff > 0 {
			gain += diff
		} else if diff < 0 {
			loss += (diff * -1.0)
		}
	}
	return gain, loss
}
