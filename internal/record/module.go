package record

import (
	"go.uber.org/fx"

	"RoboPanel-backend/internal/config"
	"RoboPanel-backend/internal/record/repository"
	"RoboPanel-backend/internal/record/service"
)

var Module = fx.Module("record",
	fx.Provide(
		func(cfg *config.Config) repository.IRecordRepository {
			return repository.NewRecordRepository(cfg.Record.DemoDir)
		},
		service.NewRecordService,
	),
)
