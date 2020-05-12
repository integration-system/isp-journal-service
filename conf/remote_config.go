package conf

import (
	"github.com/integration-system/isp-lib/v2/structure"
)

//nolint
type RemoteConfig struct {
	BaseLogDirectory string `valid:"required~Required" schema:"Путь до хранилища логов,корневая директория, в которую будут сохраняться логи"`
	CursorLifetime   int    `valid:"required~Required" schema:"Время жизни курсора в секундах,промежуток времени, в который доступно обращение к следующей записи логов по идентификатору курсора"`
	ElasticSetting   ElasticSetting
}

type ElasticSetting struct {
	Enable bool
	Config *structure.ElasticConfiguration
}
