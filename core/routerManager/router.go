package routerManager

import (
	"encoding/json"
	"errors"
	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-archaius/core/config-manager"
	"github.com/ServiceComb/go-archaius/core/event-system"
	"github.com/ServiceComb/go-chassis/core/archaius"
	"github.com/ServiceComb/go-chassis/core/config"
	"github.com/ServiceComb/go-chassis/core/lager"
	"sync"
)

const (
	RouterFileSourceName             = "RouterFileSource"
	RouterDarkLaunchGovernSourceName = "RouterDarkLaunchGovernSource"
)

const (
	DarkLaunchPrefix = "cse.darklaunch.policy."
)

var RouterRuleMgr core.ConfigMgr
var DarkLaunchGovernSource *RouterDarkLaunchGovernSource

type RouterEventListerner struct{}

func (r *RouterEventListerner) Event(event *core.Event) {
}

type RouterFileSource struct {
	once sync.Once
	d    map[string]interface{}
}

func (r *RouterFileSource) Init() {
	routerConfigs := config.GetRouterConfig()
	d := make(map[string]interface{}, 0)
	if routerConfigs == nil {
		r.d = d
		lager.Logger.Error("Can not get any router config", nil)
		return
	}
	for k, v := range routerConfigs.Destinations {
		d[k] = v
	}
	r.d = d
}

func (r *RouterFileSource) GetSourceName() string {
	return RouterFileSourceName
}
func (r *RouterFileSource) GetConfigurations() (map[string]interface{}, error) {
	r.once.Do(r.Init)
	return r.d, nil
}
func (r *RouterFileSource) GetConfigurationsByDI(dimensionInfo string) (map[string]interface{}, error) {
	return nil, nil
}
func (r *RouterFileSource) GetConfigurationByKey(k string) (interface{}, error) {
	r.once.Do(r.Init)
	v, ok := r.d[k]
	if !ok {
		return nil, errors.New("key " + k + " not exist")
	}
	return v, nil
}
func (r *RouterFileSource) GetConfigurationByKeyAndDimensionInfo(key, dimensionInfo string) (interface{}, error) {
	return nil, nil
}
func (r *RouterFileSource) AddDimensionInfo(dimensionInfo string) (map[string]string, error) {
	return nil, nil
}
func (r *RouterFileSource) DynamicConfigHandler(core.DynamicConfigCallback) error {
	return nil
}
func (r *RouterFileSource) GetPriority() int { return 10 }
func (r *RouterFileSource) Cleanup() error   { return nil }

type RouterDarkLaunchGovernSource struct {
	d core.DynamicConfigCallback
}

func (r *RouterDarkLaunchGovernSource) GetSourceName() string {
	return RouterDarkLaunchGovernSourceName
}
func (r *RouterDarkLaunchGovernSource) GetConfigurations() (map[string]interface{}, error) {
	routerConfigs, err := config.GetRouterConfigFromDarkLaunch()
	if err != nil {
		lager.Logger.Error("Get router config from dark launch failed", err)
		return nil, err
	}
	d := make(map[string]interface{}, 0)
	for k, v := range routerConfigs.Destinations {
		d[k] = v
	}
	return d, nil
}
func (r *RouterDarkLaunchGovernSource) GetConfigurationsByDI(dimensionInfo string) (map[string]interface{}, error) {
	return nil, nil
}
func (r *RouterDarkLaunchGovernSource) GetConfigurationByKey(k string) (interface{}, error) {
	s := archaius.GetString(DarkLaunchPrefix+k, "")
	rule := &config.DarkLaunchRule{}
	if err := json.Unmarshal([]byte(s), rule); err != nil {
		return nil, err
	}
	routeRules := config.TranslateRules(rule)
	return routeRules, nil
}
func (r *RouterDarkLaunchGovernSource) GetConfigurationByKeyAndDimensionInfo(key, dimensionInfo string) (interface{}, error) {
	return nil, nil
}
func (r *RouterDarkLaunchGovernSource) AddDimensionInfo(dimensionInfo string) (map[string]string, error) {
	return nil, nil
}
func (r *RouterDarkLaunchGovernSource) DynamicConfigHandler(d core.DynamicConfigCallback) error {
	r.d = d
	return nil
}
func (r *RouterDarkLaunchGovernSource) GetPriority() int { return 9 }
func (r *RouterDarkLaunchGovernSource) Cleanup() error   { return nil }
func (r *RouterDarkLaunchGovernSource) Callback(e *core.Event) error {
	if r.d == nil {
		return errors.New("dynamic config handler is nil")
	}
	lager.Logger.Infof("Router changed by dark launch governance, key: %s", e.Key)
	r.d.OnEvent(e)
	return nil
}

func Init() {
	d := eventsystem.NewDispatcher()
	d.RegisterListener(&RouterEventListerner{})
	RouterRuleMgr = configmanager.NewConfigurationManager(d)
	fileSource := &RouterFileSource{}
	RouterRuleMgr.AddSource(fileSource, fileSource.GetPriority())
	DarkLaunchGovernSource = &RouterDarkLaunchGovernSource{}
	RouterRuleMgr.AddSource(DarkLaunchGovernSource, DarkLaunchGovernSource.GetPriority())
	lager.Logger.Info("Route rule manager init success")
}

//// Refresh refresh the whole router rule config
//func Refresh() error {
//	configs := RouterConfMgr.GetConfigurations()
//	dests := make(map[string][]*config.RouteRule)
//	for k, v := range configs {
//		rules, ok := v.([]*config.RouteRule)
//		if !ok {
//			err := fmt.Errorf("route rule assertion fail, key: %s", k)
//			lager.Logger.Error(err.Error(), nil)
//			return err
//		}
//		dests[k] = rules
//	}
//	router.SetRouteRule(dests)
//	return nil
//}

// add source

// source1 local dark launch
// source2 local router
// source3 governance dark launch
