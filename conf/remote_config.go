package conf

type RemoteConfig struct {
	BaseLogDirectory string `valid:"required~Required" schema:"Путь до хранилища логов,корневая директория, в которую будут сохраняться логи"`
}
