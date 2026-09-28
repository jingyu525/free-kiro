package cli

import (
	"github.com/liujingyu/free-kiro/internal/models"
	"github.com/liujingyu/free-kiro/internal/spec"
)

// loadMetaViaEngine reads .meta.json via the engine's workspace. Used by
// subcommands that already have an Engine handle.
func loadMetaViaEngine(eng *spec.Engine, name string) (*models.SpecMeta, error) {
	return models.LoadSpecMeta(eng.WS().SpecDir(name))
}