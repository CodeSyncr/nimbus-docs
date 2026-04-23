/*
|--------------------------------------------------------------------------
| Livewire Configuration
|--------------------------------------------------------------------------
|
| This mirrors Laravel Livewire's config/livewire.php shape, adapted for Nimbus.
| Values are defaults; simple scalars can be overridden via config store keys.
|
*/

package config

import nimbusconfig "github.com/CodeSyncr/nimbus/config"

var Livewire LivewireConfig

type LivewireConfig struct {
	ComponentLocations  []string
	ComponentNamespaces map[string]string

	ComponentLayout      string
	ComponentPlaceholder string

	MakeCommand MakeCommandConfig

	ClassNamespace string
	ClassPath      string
	ViewPath       string

	TemporaryFileUpload TemporaryFileUploadConfig

	RenderOnRedirect    bool
	LegacyModelBinding  bool
	InjectAssets        bool
	Navigate            NavigateConfig
	InjectMorphMarkers  bool
	SmartWireKeys       bool
	PaginationTheme     string
	ReleaseToken        string
	CSPSafe             bool
	Payload             PayloadConfig
}

type MakeCommandConfig struct {
	Type  string // sfc | mfc | class
	Emoji bool
	With  MakeCommandWithConfig
}

type MakeCommandWithConfig struct {
	JS   bool
	CSS  bool
	Test bool
}

type TemporaryFileUploadConfig struct {
	Disk          string
	Rules         []string
	Directory     string
	Middleware    string
	PreviewMimes  []string
	MaxUploadTime int
	Cleanup       bool
}

type NavigateConfig struct {
	ShowProgressBar bool
	ProgressBarColor string
}

type PayloadConfig struct {
	MaxSize         int
	MaxNestingDepth int
	MaxCalls        int
	MaxComponents   int
}

func loadLivewire() {
	Livewire = LivewireConfig{
		ComponentLocations: []string{
			"resources/views/components",
			"resources/views/livewire",
		},
		ComponentNamespaces: map[string]string{
			"layouts": "resources/views/layouts",
			"pages":   "resources/views/pages",
		},

		ComponentLayout:      cfg("livewire.component_layout", "layouts::app"),
		ComponentPlaceholder: cfg("livewire.component_placeholder", ""),

		MakeCommand: MakeCommandConfig{
			Type:  cfg("livewire.make_command.type", "sfc"),
			Emoji: nimbusconfig.GetOrDefault("livewire.make_command.emoji", true),
			With: MakeCommandWithConfig{
				JS:   nimbusconfig.GetOrDefault("livewire.make_command.with.js", false),
				CSS:  nimbusconfig.GetOrDefault("livewire.make_command.with.css", false),
				Test: nimbusconfig.GetOrDefault("livewire.make_command.with.test", false),
			},
		},

		ClassNamespace: cfg("livewire.class_namespace", "App\\Livewire"),
		ClassPath:      cfg("livewire.class_path", "app/Livewire"),
		ViewPath:       cfg("livewire.view_path", "resources/views/livewire"),

		TemporaryFileUpload: TemporaryFileUploadConfig{
			Disk:       cfg("livewire.temporary_file_upload.disk", ""),
			Rules:      nil,
			Directory:  cfg("livewire.temporary_file_upload.directory", ""),
			Middleware: cfg("livewire.temporary_file_upload.middleware", ""),
			PreviewMimes: []string{
				"png", "gif", "bmp", "svg", "wav", "mp4",
				"mov", "avi", "wmv", "mp3", "m4a",
				"jpg", "jpeg", "mpga", "webp", "wma",
			},
			MaxUploadTime: nimbusconfig.GetOrDefault("livewire.temporary_file_upload.max_upload_time", 5),
			Cleanup:       nimbusconfig.GetOrDefault("livewire.temporary_file_upload.cleanup", true),
		},

		RenderOnRedirect:   nimbusconfig.GetOrDefault("livewire.render_on_redirect", false),
		LegacyModelBinding: nimbusconfig.GetOrDefault("livewire.legacy_model_binding", false),
		InjectAssets:       nimbusconfig.GetOrDefault("livewire.inject_assets", true),

		Navigate: NavigateConfig{
			ShowProgressBar: nimbusconfig.GetOrDefault("livewire.navigate.show_progress_bar", true),
			ProgressBarColor: cfg("livewire.navigate.progress_bar_color", "#2299dd"),
		},

		InjectMorphMarkers: nimbusconfig.GetOrDefault("livewire.inject_morph_markers", true),
		SmartWireKeys:      nimbusconfig.GetOrDefault("livewire.smart_wire_keys", true),
		PaginationTheme:    cfg("livewire.pagination_theme", "tailwind"),
		ReleaseToken:       cfg("livewire.release_token", "a"),
		CSPSafe:            nimbusconfig.GetOrDefault("livewire.csp_safe", false),

		Payload: PayloadConfig{
			MaxSize:         nimbusconfig.GetOrDefault("livewire.payload.max_size", 1024*1024),
			MaxNestingDepth: nimbusconfig.GetOrDefault("livewire.payload.max_nesting_depth", 10),
			MaxCalls:        nimbusconfig.GetOrDefault("livewire.payload.max_calls", 50),
			MaxComponents:   nimbusconfig.GetOrDefault("livewire.payload.max_components", 20),
		},
	}
}

