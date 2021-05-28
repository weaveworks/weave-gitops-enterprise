package templates

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRenderTemplate(t *testing.T) {
	testTemplate := []byte("value: ${TEST_VALUE}\n")
	parameters := TemplateParams{
		"TEST_VALUE": "foo",
	}

	fakeLib := fakeLibrary{
		templates: map[string][]byte{
			"aws-fargate-eks": testTemplate,
		},
	}

	b, err := RenderTemplate(context.TODO(), fakeLib, "aws-fargate-eks", parameters)
	fatalIfError(t, err)

	if diff := cmp.Diff("value: foo\n", string(b)); diff != "" {
		t.Fatalf("failed to render template:\n%s", diff)
	}
}

func TestRenderTemplate_failed_to_get_template(t *testing.T) {
	parameters := TemplateParams{
		"TEST_VALUE": "foo",
	}

	fakeLib := fakeLibrary{
		templates: map[string][]byte{},
	}

	_, err := RenderTemplate(context.TODO(), fakeLib, "aws-fargate-eks", parameters)
	assertErrorMatch(t, "could not find template \"aws-fargate-eks\": unknown template", err)
}

func Test_renderTemplate(t *testing.T) {
	testTemplate := []byte("value: ${TEST_VALUE}\n")
	parameters := TemplateParams{
		"TEST_VALUE": "foo",
	}

	b, err := renderTemplate(testTemplate, parameters)
	fatalIfError(t, err)

	if diff := cmp.Diff("value: foo\n", string(b)); diff != "" {
		t.Fatalf("failed to render template:\n%s", diff)
	}
}

func TestRenderTemplate_missing_param(t *testing.T) {
	testTemplate := []byte("value: ${TEST_VALUE}\n")
	parameters := TemplateParams{}

	_, err := renderTemplate(testTemplate, parameters)
	if err == nil {
		t.Fatal("expected to get an error rendering the template")
	}
	assertErrorMatch(t, "failed to render template:.*value for variable.*TEST_VALUE.*is not set", err)
}

func assertErrorMatch(t *testing.T, s string, e error) {
	t.Helper()
	if s == "" && e == nil {
		return
	}
	if s != "" && e == nil {
		t.Fatalf("expected error %s, got nil", s)
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Fatalf("got error %s, want %s", s, e)
	}
}

func fatalIfError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type fakeLibrary struct {
	templates map[string][]byte
}

func (f fakeLibrary) Get(ctx context.Context, name string) ([]byte, error) {
	if b, ok := f.templates[name]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("unknown template: %s", name)
}
