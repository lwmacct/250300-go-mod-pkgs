package query_range

// path: $.
type Root struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

// path: $.data
type Data struct {
	ResultType string       `json:"resultType"`
	Result     []ResultItem `json:"result"`
	Stats      Stats        `json:"stats"`
}

// path: $.data.result
type ResultItem struct {
	Stream map[string]string `json:"stream" note:"这是数据包含的 label"`
	Values [][]string        `json:"values" note:"数组成员=2, 第一个是时间戳, 第二个是值"`
}

// path: $.data.stats
type Stats struct {
	Summary  SummaryStats  `json:"summary"`
	Querier  QuerierStats  `json:"querier"`
	Ingester IngesterStats `json:"ingester"`
	Cache    CacheStats    `json:"cache"`
	Index    IndexStats    `json:"index"`
}

// path: $.data.stats.summary
type SummaryStats struct {
	BytesProcessedPerSecond               int     `json:"bytesProcessedPerSecond"`
	LinesProcessedPerSecond               int     `json:"linesProcessedPerSecond"`
	TotalBytesProcessed                   int     `json:"totalBytesProcessed"`
	TotalLinesProcessed                   int     `json:"totalLinesProcessed"`
	ExecTime                              float64 `json:"execTime"`
	QueueTime                             float64 `json:"queueTime"`
	Subqueries                            int     `json:"subqueries"`
	TotalEntriesReturned                  int     `json:"totalEntriesReturned"`
	Splits                                int     `json:"splits"`
	Shards                                int     `json:"shards"`
	TotalPostFilterLines                  int     `json:"totalPostFilterLines"`
	TotalStructuredMetadataBytesProcessed int     `json:"totalStructuredMetadataBytesProcessed"`
}

// path: $.data.stats.querier
type QuerierStats struct {
	Store StoreStats `json:"store"`
}

// path: $.data.stats.ingester
type IngesterStats struct {
	TotalReached       int        `json:"totalReached"`
	TotalChunksMatched int        `json:"totalChunksMatched"`
	TotalBatches       int        `json:"totalBatches"`
	TotalLinesSent     int        `json:"totalLinesSent"`
	Store              StoreStats `json:"store"`
}

// path: $.data.stats.ingester.store
type StoreStats struct {
	TotalChunksRef                    int        `json:"totalChunksRef"`
	TotalChunksDownloaded             int        `json:"totalChunksDownloaded"`
	ChunksDownloadTime                int        `json:"chunksDownloadTime"`
	QueryReferencedStructuredMetadata bool       `json:"queryReferencedStructuredMetadata"`
	Chunk                             ChunkStats `json:"chunk"`
	ChunkRefsFetchTime                int        `json:"chunkRefsFetchTime"`
	CongestionControlLatency          int        `json:"congestionControlLatency"`
	PipelineWrapperFilteredLines      int        `json:"pipelineWrapperFilteredLines"`
}

// path: $.data.stats.ingester.store.chunk
type ChunkStats struct {
	HeadChunkBytes                      int `json:"headChunkBytes"`
	HeadChunkLines                      int `json:"headChunkLines"`
	DecompressedBytes                   int `json:"decompressedBytes"`
	DecompressedLines                   int `json:"decompressedLines"`
	CompressedBytes                     int `json:"compressedBytes"`
	TotalDuplicates                     int `json:"totalDuplicates"`
	PostFilterLines                     int `json:"postFilterLines"`
	HeadChunkStructuredMetadataBytes    int `json:"headChunkStructuredMetadataBytes"`
	DecompressedStructuredMetadataBytes int `json:"decompressedStructuredMetadataBytes"`
}

// path: $.data.stats.cache
type CacheStats struct {
	Chunk               CacheSubStats `json:"chunk"`
	Index               CacheSubStats `json:"index"`
	Result              CacheSubStats `json:"result"`
	StatsResult         CacheSubStats `json:"statsResult"`
	VolumeResult        CacheSubStats `json:"volumeResult"`
	SeriesResult        CacheSubStats `json:"seriesResult"`
	LabelResult         CacheSubStats `json:"labelResult"`
	InstantMetricResult CacheSubStats `json:"instantMetricResult"`
}

// path: $.data.stats.cache.*
type CacheSubStats struct {
	EntriesFound      int `json:"entriesFound"`
	EntriesRequested  int `json:"entriesRequested"`
	EntriesStored     int `json:"entriesStored"`
	BytesReceived     int `json:"bytesReceived"`
	BytesSent         int `json:"bytesSent"`
	Requests          int `json:"requests"`
	DownloadTime      int `json:"downloadTime"`
	QueryLengthServed int `json:"queryLengthServed"`
}

// path: $.data.stats.index
type IndexStats struct {
	TotalChunks      int  `json:"totalChunks"`
	PostFilterChunks int  `json:"postFilterChunks"`
	ShardsDuration   int  `json:"shardsDuration"`
	UsedBloomFilters bool `json:"usedBloomFilters"`
}
