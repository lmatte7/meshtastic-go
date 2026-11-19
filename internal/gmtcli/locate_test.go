package gmtcli

import (
	"math"
	"testing"
	"time"
)

func TestEstimateLocationPrefersGPS(t *testing.T) {
	nodes := []MeshNode{
		{NodeNumber: 1, LatitudeDegrees: 36.1, LongitudeDegrees: -115.2, AltitudeMeters: 620},
		{NodeNumber: 2, LatitudeDegrees: 37, LongitudeDegrees: -116, SNR: 10, LastHeard: "t", LastHeardSecondsAgo: 60},
	}
	loc, err := estimateLocation(1, nodes, locateMaxAge)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Source != "gps" || loc.LatitudeDegrees != 36.1 || loc.LongitudeDegrees != -115.2 || loc.AltitudeMeters == nil || *loc.AltitudeMeters != 620 {
		t.Errorf("location = %+v", loc)
	}
}

func TestEstimateLocationWeightsBySNR(t *testing.T) {
	// 10 dB difference is a factor of ten in power, so the strong node gets
	// ten times the weight of the weak one.
	nodes := []MeshNode{
		{NodeNumber: 1}, // our radio, no fix
		{NodeNumber: 2, ID: "!2", LatitudeDegrees: 36.0, LongitudeDegrees: -115.0, SNR: 10, LastHeard: "t", LastHeardSecondsAgo: 60},
		{NodeNumber: 3, ID: "!3", LatitudeDegrees: 36.11, LongitudeDegrees: -115.11, SNR: 0, LastHeard: "t", LastHeardSecondsAgo: 120},
	}
	loc, err := estimateLocation(1, nodes, locateMaxAge)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Source != "estimate" || len(loc.Nodes) != 2 {
		t.Fatalf("location = %+v", loc)
	}
	// (10*36.0 + 1*36.11) / 11 = 36.01
	if math.Abs(loc.LatitudeDegrees-36.01) > 1e-9 || math.Abs(loc.LongitudeDegrees-(-115.01)) > 1e-9 {
		t.Errorf("estimate = %v, %v; want 36.01, -115.01", loc.LatitudeDegrees, loc.LongitudeDegrees)
	}
	if loc.Nodes[0].ID != "!2" || loc.Nodes[0].Weight != 0.909 || loc.Nodes[1].Weight != 0.091 {
		t.Errorf("weights = %+v", loc.Nodes)
	}
}

func TestEstimateLocationFilters(t *testing.T) {
	good := MeshNode{NodeNumber: 2, LatitudeDegrees: 36, LongitudeDegrees: -115, SNR: 5, LastHeard: "t", LastHeardSecondsAgo: 60}
	cases := map[string]MeshNode{
		"via mqtt":         {NodeNumber: 3, LatitudeDegrees: 40, LongitudeDegrees: -100, SNR: 20, LastHeard: "t", LastHeardSecondsAgo: 10, ViaMQTT: true},
		"relayed":          {NodeNumber: 4, LatitudeDegrees: 40, LongitudeDegrees: -100, SNR: 20, LastHeard: "t", LastHeardSecondsAgo: 10, HopsAway: 1},
		"too old":          {NodeNumber: 5, LatitudeDegrees: 40, LongitudeDegrees: -100, SNR: 20, LastHeard: "t", LastHeardSecondsAgo: int(locateMaxAge/time.Second) + 1},
		"never heard":      {NodeNumber: 6, LatitudeDegrees: 40, LongitudeDegrees: -100, SNR: 20},
		"without position": {NodeNumber: 7, SNR: 20, LastHeard: "t", LastHeardSecondsAgo: 10},
	}
	for name, excluded := range cases {
		loc, err := estimateLocation(1, []MeshNode{{NodeNumber: 1}, good, excluded}, locateMaxAge)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(loc.Nodes) != 1 || loc.LatitudeDegrees != 36 || loc.LongitudeDegrees != -115 {
			t.Errorf("%s: node should be excluded, got %+v", name, loc)
		}
	}

	if _, err := estimateLocation(1, []MeshNode{{NodeNumber: 1}}, locateMaxAge); err == nil {
		t.Error("expected an error with no fix and no usable nodes")
	}
}
