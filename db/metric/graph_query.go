package metric

const (
	EndPointAddress     = "address"
	EndPointAddresses   = "addresses"
	EndPointBlock       = "block"
	EndPointBlocks      = "blocks"
	EndPointBlockNewest = "block_newest"
	EndPointPosts       = "posts"
	EndPointPostsNewest = "posts_newest"
	EndPointProfiles    = "profiles"
	EndPointRoom        = "room"
	EndPointTx          = "tx"
)

func AddGraphQuery(endpoint string) {
	writer := getInfluxWriter()
	if writer == nil {
		return
	}
	writer.Write(Point{
		Measurement: NameGraphQuery,
		Fields: map[string]interface{}{
			FieldQuantity: 1,
		},
		Tags: map[string]string{
			TagEndpoint: endpoint,
		},
	})
}
