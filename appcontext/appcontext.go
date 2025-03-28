package appcontext

import (
	"CodeAutoGo/cmdclient"
	"CodeAutoGo/config"
	"sync"
)

type AppContext struct {
	GitClient    *cmdclient.GitClient
	CodeQLClient *cmdclient.CodeQLClient
	Config       *config.Config
	TaskStatus   sync.Map
}
