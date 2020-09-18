package main

import (
	"time"
)

func makeGraphiteAnswer(victoriaMetricsAnswer *victoriaMetricsAnswerStruct) []graphiteStruct {

	graphiteAnswer := make([]graphiteStruct, len(victoriaMetricsAnswer.Metrics))
	// graphiteIntervalsVar := make(graphiteIntervals, len(victoriaMetricsAnswer.Metrics))

	for i := range victoriaMetricsAnswer.Metrics {
		graphiteAnswer[i].Path = victoriaMetricsAnswer.Metrics[i].Path

		if victoriaMetricsAnswer.Metrics[i].IsLeaf == 0 {
			graphiteAnswer[i].IsLeaf = false
			continue
		}
		g := make(graphiteIntervals, 1)
		gSliceTwo := make([]int, 2)
		gSliceTwo[0] = 0
		gSliceTwo[1] = int(time.Now().UnixNano())
		g[0] = gSliceTwo
		graphiteAnswer[i].IsLeaf = true
		graphiteAnswer[i].Intervals = g
		/*
			graphiteIntervalsData[0][0] = 0
			graphiteIntervalsData[0][1] = int(time.Now().UnixNano())
			graphiteAnswer[i].Intervals = graphiteIntervalsData
		*/
	}

	return graphiteAnswer

}
