package routerManager

import (
	"fmt"

	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-archaius/core/config-manager"
	"github.com/ServiceComb/go-archaius/core/event-system"
	"github.com/ServiceComb/go-chassis/core/config"
	"github.com/ServiceComb/go-chassis/core/lager"
	"github.com/ServiceComb/go-chassis/core/route"
)

var RouteConfMgr core.ConfigMgr

func Init() {
	d := eventsystem.NewDispatcher()
	RouteConfMgr = configmanager.NewConfigurationManager(d)
}

// Refresh refresh the whole router rule config
func Refresh() error {
	configs := RouteConfMgr.GetConfigurations()
	dests := make(map[string][]*config.RouteRule)
	for k, v := range configs {
		rules, ok := v.([]*config.RouteRule)
		if !ok {
			err := fmt.Errorf("route rule assertion fail, key: %s", k)
			lager.Logger.Error(err.Error(), nil)
			return err
		}
		dests[k] = rules
	}
	router.SetRouteRule(dests)
	return nil
}

// add source
// source1 local dark launch
// source2 local router
// source3 governance dark launch
// source4 governance router
// register listener
