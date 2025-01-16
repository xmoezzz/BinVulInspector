package archive

import "testing"

func TestCompressor_Archive(t *testing.T) {
	files := map[string]string{
		"E:\\other\\test\\repo\\macos_java": "macos_java",
	}

	err := NewCompressor().Archive("E:\\other\\test\\repo\\macos_java.zip", files, WithMaxFileCountRule(100), WithMaxFileSizeRule(100<<20))
	if err != nil {
		t.Error(err)
	}

	t.Log("success")
}

func TestCompressor_Extract(t *testing.T) {
	err := NewCompressor().Extract("E:\\other\\test\\repo\\extract", "E:\\other\\test\\repo\\macos_java.zip", WithMaxFileCountRule(81), WithMaxFileSizeRule(100<<20))
	if err != nil {
		t.Error(err)
	}

	t.Log("success")
}
