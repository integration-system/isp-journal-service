package helper

import (
	"github.com/integration-system/isp-lib/v2/structure"
	"isp-journal-service/controller"
)

func GetAllEndpoints(moduleName string) []structure.EndpointDescriptor {
	return structure.DescriptorsWithPrefix(moduleName, []structure.EndpointDescriptor{
		{
			Path:    "log/search",
			Handler: controller.SearchController.Search,
			Inner:   true,
		},
		{
			Path:    "log/search_with_cursor",
			Handler: controller.SearchController.SearchWithCursor,
			Inner:   true,
		},
		{
			Path:    "log/export",
			Handler: controller.ExportController.Export,
			Inner:   true,
		},
		{
			Path:    "log/transfer",
			Handler: controller.LogController.Transfer,
			Inner:   true,
		},
	})
}
