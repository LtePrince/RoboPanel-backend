package record

import (
	"go.uber.org/fx"

	"robot-panel/internal/record/service"
)

var Module = fx.Module("record",
	fx.Provide(service.NewRecordService),
)
