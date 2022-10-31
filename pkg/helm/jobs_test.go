package helm

import "testing"

func TestJobs(t *testing.T) {
	// should just keep incrementing nicely
	j := NewJobs()
	a := j.New()
	b := j.New()

	aData, found := j.Get(a)
	if found != true {
		t.Errorf("expected to find job %s", a)
	}
	if aData.Result != "" {
		t.Errorf("expected empty result, got %s", aData.Result)
	}

	j.Set(a, JobResult{Result: "foo"})
	aData2, _ := j.Get(a)
	if aData2.Result != "foo" {
		t.Errorf("expected result foo, got %s", aData2.Result)
	}

	// bshould still be empty
	bData, found := j.Get(b)
	if found != true {
		t.Errorf("expected to find job %s", b)
	}
	if bData.Result != "" {
		t.Errorf("expected empty result, got %s", bData.Result)
	}

	cData, found := j.Get("foo")
	if found != false {
		t.Errorf("expected not to find id 3, got %s", cData.Result)
	}
}
