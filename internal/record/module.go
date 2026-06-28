package record

import (
	"go.uber.org/fx"

	"robot-panel/internal/config"
	"robot-panel/internal/record/repository"
	"robot-panel/internal/record/service"
)

var Module = fx.Module("record",
	fx.Provide(
		func(cfg *config.Config) repository.IRecordRepository {
			return repository.NewRecordRepository(cfg.Record.DemoDir)
		},
		service.NewRecordService,
	),
)
