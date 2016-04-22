package render_test

import (
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/render/expected"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestHostRenderer(t *testing.T) {
	have := render.HostRenderer.Render(fixture.Report)
	want := expected.RenderedHosts
	if !RoughlyEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}
