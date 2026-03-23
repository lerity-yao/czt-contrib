package rabbitmq

import (
	"fmt"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
)

func generateRabbitmqEtcNames(api *spec.ApiSpec) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, g := range api.Service.Groups {
		for _, h := range g.Routes {
			l := fmt.Sprintf(
				"%sRabbitmqConf: \n  Username:\n  Password:\n  Host:\n  Port:\n  ListenerQueues:\n    - Name: %s\n",
				strings.TrimSuffix(h.Handler, "Handler"), strings.TrimPrefix(h.Path, "/"))
			if _, ok := seen[l]; !ok {
				seen[l] = struct{}{}
				names = append(names, l)
			}
		}
	}
	return names
}

func generateRabbitmqConfigNames(api *spec.ApiSpec) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, g := range api.Service.Groups {
		for _, h := range g.Routes {
			l := fmt.Sprintf("%sRabbitmqConf rabbitmq.RabbitListenerConf", strings.TrimSuffix(h.Handler, "Handler"))
			if _, ok := seen[l]; !ok {
				seen[l] = struct{}{}
				names = append(names, l)
			}
		}
	}
	return names
}

// GetDoc formats a doc string with comment prefix.
func GetDoc(doc string) string {
	if len(doc) == 0 {
		return ""
	}

	return "// " + strings.Trim(doc, "\"")
}
