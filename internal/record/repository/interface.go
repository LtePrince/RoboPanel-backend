package repository

import "RoboPanel-backend/internal/record/schema"

type IRecordRepository interface {
	ListDemos() ([]schema.Demo, error)
	FileExists(demoName, fileName string) (string, bool)
}
