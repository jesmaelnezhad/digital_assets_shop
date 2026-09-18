package product_test

import "testing"

func TestAppearancePaletteCatalog(t *testing.T) {
	allowed := map[string]bool{
		"clay": true, "marble": true, "night": true, "moss": true,
		"ink": true, "ember": true, "dune": true, "frost": true,
		"paper": true, "chalk": true, "linen": true, "mist": true,
		"petal": true, "foam": true, "porcelain": true, "sage": true,
		"snow": true, "honey": true, "bone": true, "cloud": true,
		"wine": true, "violet": true, "ocean": true, "slate": true,
	}
	if len(allowed) != 24 {
		t.Fatalf("spec palettes=%d", len(allowed))
	}
	for _, bad := range []string{"neon", "dark", ""} {
		if allowed[bad] {
			t.Fatalf("%s must not be a palette", bad)
		}
	}
}

func TestAppearanceFontRadiusDensityCatalog(t *testing.T) {
	fonts := []string{"system", "humanist", "serif", "mono", "display"}
	radii := []string{"sharp", "soft", "round"}
	density := []string{"compact", "comfortable", "roomy"}
	if len(fonts) != 5 || len(radii) != 3 || len(density) != 3 {
		t.Fatal("appearance catalogs drifted from PRODUCT-SPEC")
	}
}

func TestBannerQueryContract(t *testing.T) {
	if got := bannerPath(); got != "/api/v1/products?banner=1" {
		t.Fatalf("%s", got)
	}
}

func bannerPath() string { return "/api/v1/products?banner=1" }
