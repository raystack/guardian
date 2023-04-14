package provider

import (
	"fmt"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/newpoc"
)

// TODO: move this to guardian/plugins/providers
type pluginFactory struct {
	clients map[string]clientV2
	configs map[string]providerConfig
}

func (f *pluginFactory) getConfig(pc *domain.ProviderConfig) (providerConfig, error) {
	if f.configs == nil {
		f.configs = make(map[string]providerConfig)
	}

	key := pc.URN
	if config, ok := f.configs[key]; ok {
		return config, nil
	}

	switch pc.Type {
	case newpoc.ProviderType:
		config, err := newpoc.NewConfig(pc)
		if err != nil {
			return nil, err
		}
		f.configs[key] = config
		return config, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %q", pc.Type)
	}
}

func (f *pluginFactory) getClient(cfg providerConfig) (clientV2, error) {
	if f.clients == nil {
		f.clients = make(map[string]clientV2)
	}

	key := cfg.GetProviderConfig().URN
	if client, ok := f.clients[key]; ok {
		return client, nil
	}

	providerType := cfg.GetProviderConfig().Type
	switch providerType {
	case newpoc.ProviderType:
		client, err := newpoc.NewClient(cfg.(*newpoc.Config))
		if err != nil {
			return nil, err
		}
		f.clients[key] = client
		return client, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %q", providerType)
	}
}
