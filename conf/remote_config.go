package conf

type RemoteConfig struct {
	BaseLogDirectory string `valid:"required~Required"`
}
