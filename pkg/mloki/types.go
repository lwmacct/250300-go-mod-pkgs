package mloki

import (
	"github.com/lwmacct/250109-m-bdflow/module/v00/mloki/types/label_values"
	"github.com/lwmacct/250109-m-bdflow/module/v00/mloki/types/query_range"
	"github.com/lwmacct/250109-m-bdflow/module/v00/mloki/types/tail"
)

// GET /loki/api/v1/query_range
type TsQueryRange query_range.Root

// GET /loki/api/v1/label_values
type TsLabelValues label_values.Root

// GET /loki/api/v1/tail
type TsTail tail.Root
