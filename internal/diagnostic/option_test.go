package diagnostic

import "testing"

func TestOptions(t *testing.T) {
	t.Run("WithCategory", func(t *testing.T) {
		d := &diagnostic{}
		opt := WithCategory(Category_InvalidParamter)
		opt(d)

		if d.category != Category_InvalidParamter {
			t.Errorf("WithCategory failed; got %q", d.category)
		}
	})

	t.Run("WithoutStacktrace", func(t *testing.T) {
		d := &diagnostic{}
		opt := WithoutStacktrace()
		opt(d)

		if d.stacktrace == nil {
			t.Fatal("WithoutStacktrace initialized nil slice instead of empty non-nil structural type")
		}
		if len(d.stacktrace) != 0 {
			t.Errorf("WithoutStacktrace left un-cleared structural items in trace array")
		}
	})

	t.Run("WithStacktrace frame captures", func(t *testing.T) {
		d := &diagnostic{}
		WithStacktrace()(d)

		if len(d.stacktrace) == 0 {
			t.Fatal("WithStacktrace failed to automatically extract caller stacks")
		}
	})
}
