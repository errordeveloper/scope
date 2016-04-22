package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestEndpointRenderer(t *testing.T) {
	have := render.EndpointRenderer.Render(fixture.Report)
	want := expected.RenderedEndpoints
	if !RoughlyEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessRenderer(t *testing.T) {
	have := render.ProcessRenderer.Render(fixture.Report)
	want := expected.RenderedProcesses
	if !RoughlyEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestProcessNameRenderer(t *testing.T) {
	have := render.ProcessNameRenderer.Render(fixture.Report)
	want := expected.RenderedProcessNames
	if !RoughlyEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
