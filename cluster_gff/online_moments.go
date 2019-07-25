package main

import "math"

type CutPoints struct {
	Left  int
	Right int
}

type OnlineMoments struct {
	Count int
	Min   float64
	Max   float64
	Mean  float64
	Sd    float64
	m     float64
	Cut   CutPoints
}

func NewOnlineMoments() *OnlineMoments {
	return &OnlineMoments{0, math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), CutPoints{-1, -1}}
}

func (h *OnlineMoments) Update(p float64) {
	if h.Count == 0 {
		h.Min = p
		h.Max = p
		h.Mean = p
		h.Sd = 0.0
		h.m = 0.0
		h.Count++
		return
	} else {
		h.Count++
		if p < h.Min {
			h.Min = p
		}
		if p > h.Max {
			h.Max = p
		}
		oldMean := h.Mean
		h.Mean += (p - h.Mean) / float64(h.Count)
		h.m += (p - oldMean) * (p - h.Mean)
		h.Sd = math.Sqrt(h.m / float64(h.Count))
	}
}
