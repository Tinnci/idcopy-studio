package idcopy

import "testing"

func baseRequest() LayoutRequest {
	return LayoutRequest{
		PaperID: "iso-a4", DocumentID: "iso-id-1", Orientation: "portrait",
		Arrangement: "paired", Sides: "both", Copies: 2, ScalePercent: 100,
		MarginMM: 12, GapMM: 8,
	}
}

func TestPairedID1LayoutOnA4(t *testing.T) {
	plan, err := Plan(baseRequest())
	if err != nil {
		t.Fatal(err)
	}
	if plan.Pages != 1 {
		t.Fatalf("expected one page, got %d", plan.Pages)
	}
	if len(plan.Items) != 4 {
		t.Fatalf("expected four placed sides, got %d", len(plan.Items))
	}
	if plan.Items[1].X <= plan.Items[0].X {
		t.Fatal("back side should be placed to the right of front side")
	}
}

func TestAutoOrientationChoosesHigherCapacity(t *testing.T) {
	request := baseRequest()
	request.Orientation = "auto"
	request.Arrangement = "grid"
	request.Sides = "front"
	request.Copies = 20
	plan, err := Plan(request)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Orientation != "landscape" {
		t.Fatalf("expected landscape for an ID-1 grid, got %s", plan.Orientation)
	}
}

func TestTD3DoesNotFitA6WithMargins(t *testing.T) {
	request := baseRequest()
	request.PaperID = "iso-a6"
	request.DocumentID = "icao-td3"
	request.MarginMM = 12
	if _, err := Plan(request); err == nil {
		t.Fatal("expected layout failure")
	}
}

func TestGridPaginatesIndividualSides(t *testing.T) {
	request := baseRequest()
	request.PaperID = "photo-4x6"
	request.Arrangement = "grid"
	request.Orientation = "portrait"
	request.Copies = 3
	request.MarginMM = 5
	request.GapMM = 3
	plan, err := Plan(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 6 || plan.Pages != 3 {
		t.Fatalf("expected six items over three pages, got %d items over %d pages", len(plan.Items), plan.Pages)
	}
}
