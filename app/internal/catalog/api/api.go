package api

type CdcOperation string

var (
	// There's no way to create a distinction between creating and updating, in both cases "before" is nil.
	// It's true both for UPDATE and UPSERT clauses.
	CdcOperationUpsert CdcOperation = "u"
	CdcOperationDelete CdcOperation = "d"
)

type ProductChangeCdcMessage struct {
	Payload ProductChangeCdcMessagePayload `json:"payload"`
}

type ProductChangeCdcMessagePayload struct {
	Before    *ProductChangeSchema `json:"before"`
	After     *ProductChangeSchema `json:"after"`
	Operation CdcOperation         `json:"op"`
}

type ProductChangeSchema struct {
	Id                  string   `json:"id"`
	Name                *string  `json:"name"`
	Description         *string  `json:"description"`
	PicturesJsonListStr *string  `json:"pictures"`
	Price               *float64 `json:"price"`
	Stock               *uint32  `json:"stock"`
	CreatedAtUnixMs     *int64   `json:"created_at"`
	UpdatedAtUnixMs     *int64   `json:"updated_at"`
	DeletedAtUnixMs     *int64   `json:"deleted_at"`
}

type ProductChange struct {
	Id              string
	Name            string
	Description     string
	Pictures        []ProductsChangePicture
	Price           float64
	Stock           uint32
	CreatedAtUnixMs int64
	UpdatedAtUnixMs int64
	DeletedAtUnixMs *int64
}
type ProductsChangePicture struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type DataStreamProductChangeCdcMessages struct {
	Messages []ProductChangeCdcMessage `json:"messages"`
}
