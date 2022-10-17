package helm

import "testing"

func TestJobs(t *testing.T) {
	// should just keep incrementing nicely
	j := NewJobs()
	a := j.New()
	if a != "1" {
		t.Errorf("expected id 1, got %s", a)
	}
	b := j.New()
	if b != "2" {
		t.Errorf("expected id 2, got %s", b)
	}

	aData, found := j.Get("1")
	if found != true {
		t.Errorf("expected to find id 1")
	}
	if aData.Result != "" {
		t.Errorf("expected empty result, got %s", aData.Result)
	}

	j.Set("1", JobResult{Result: "foo"})
	aData2, _ := j.Get("1")
	if aData2.Result != "foo" {
		t.Errorf("expected result foo, got %s", aData2.Result)
	}

	// "2" should still be empty
	bData, found := j.Get("2")
	if found != true {
		t.Errorf("expected to find id 2")
	}
	if bData.Result != "" {
		t.Errorf("expected empty result, got %s", bData.Result)
	}

	cData, found := j.Get("3")
	if found != false {
		t.Errorf("expected not to find id 3, got %s", cData.Result)
	}
}
