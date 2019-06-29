// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package app

import (
	"code.cloudfoundry.org/clock"
	"github.com/cloudfoundry/bosh-agent/agent"
	"github.com/cloudfoundry/bosh-agent/agent/applier/applyspec"
	"github.com/cloudfoundry/bosh-agent/agent/bootonce"
	"github.com/cloudfoundry/bosh-agent/handler"
	"github.com/cloudfoundry/bosh-agent/infrastructure"
	"github.com/cloudfoundry/bosh-agent/jobsupervisor"
	"github.com/cloudfoundry/bosh-agent/platform"
	"github.com/cloudfoundry/bosh-agent/settings"
	"github.com/cloudfoundry/bosh-agent/settings/directories"
	"github.com/cloudfoundry/bosh-agent/sigar"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
	"github.com/cloudfoundry/bosh-utils/uuid"
	sigar2 "github.com/cloudfoundry/gosigar"
	"time"
)

// Injectors from wire.go:

func InitializeDirProvider(baseDir string) directories.Provider {
	provider := directories.NewProvider(baseDir)
	return provider
}

func InitializeAuditLogger(logger2 logger.Logger) *platform.DelayedAuditLogger {
	auditLoggerProvider := platform.NewAuditLoggerProvider()
	delayedAuditLogger := platform.NewDelayedAuditLogger(auditLoggerProvider, logger2)
	return delayedAuditLogger
}

func NewPlatform(logger2 logger.Logger, dirProvider directories.Provider, fs system.FileSystem, opts platform.Options, state *platform.BootstrapState, clock2 clock.Clock, auditLogger platform.AuditLogger, name string) (platform.Platform, error) {
	sigarSigar := NewConcreteSigar()
	collector := sigar.NewSigarStatsCollector(sigarSigar)
	provider := platform.NewProvider(logger2, dirProvider, collector, fs, opts, state, clock2, auditLogger)
	platformPlatform, err := ProvidePlatform(provider, name)
	if err != nil {
		return nil, err
	}
	return platformPlatform, nil
}

func InitializeSettingsSourceFactory(opts infrastructure.SettingsOptions, platform2 platform.Platform, logger2 logger.Logger) infrastructure.SettingsSourceFactory {
	settingsSourceFactory := infrastructure.NewSettingsSourceFactory(opts, platform2, logger2)
	return settingsSourceFactory
}

func InitializeAgent(logger2 logger.Logger, mbusHandler handler.Handler, platform2 platform.Platform, actionDispatcher agent.ActionDispatcher, jobSupervisor jobsupervisor.JobSupervisor, specService applyspec.V1Service, heartbeatInterval time.Duration, settingsService settings.Service, uuidGenerator uuid.Generator, timeService clock.Clock, dirProvider directories.Provider, fs system.FileSystem) agent.Agent {
	startManager := NewStartManager(settingsService, fs, dirProvider)
	agentAgent := agent.New(logger2, mbusHandler, platform2, actionDispatcher, jobSupervisor, specService, heartbeatInterval, settingsService, uuidGenerator, timeService, startManager)
	return agentAgent
}

// wire.go:

func NewConcreteSigar() sigar2.Sigar {
	return &sigar2.ConcreteSigar{}
}

func ProvidePlatform(p platform.Provider, name string) (platform.Platform, error) {
	return p.Get(name)
}

func NewStartManager(settingsService settings.Service, fs system.FileSystem, dirProvider directories.Provider) agent.StartManager {
	return bootonce.NewStartManager(
		settingsService,
		fs,
		dirProvider,
	)
}
