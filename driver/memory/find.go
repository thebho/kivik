package memory

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/util"
)

var errFindNotImplemented = errors.New("find feature not yet implemented")

type findQuery struct {
	Selector map[string]interface{} `json:"selector"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`
	Sort     []string               `json:"sort"`
	Fields   []string               `json:"fields"`
	UseIndex indexSpec              `json:"use_index"`
}

type indexSpec struct {
	ddoc  string
	index string
}

func (i *indexSpec) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &i.ddoc)
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if len(values) == 0 || len(values) > 2 {
		return errors.New("invalid index specification")
	}
	i.ddoc = values[0]
	if len(values) == 2 {
		i.index = values[1]
	}
	return nil
}

func (d *db) CreateIndex(_ context.Context, ddoc, name string, index interface{}) error {
	return errFindNotImplemented
}

func (d *db) GetIndexes(_ context.Context) ([]driver.Index, error) {
	return nil, errFindNotImplemented
}

func (d *db) DeleteIndex(_ context.Context, ddoc, name string) error {
	return errFindNotImplemented
}

func (d *db) Find(_ context.Context, query interface{}) (driver.Rows, error) {
	queryJSON, err := util.ToJSON(query)
	if err != nil {
		return nil, err
	}
	fq := &findQuery{}
	if err := json.NewDecoder(queryJSON).Decode(&fq); err != nil {
		return nil, err
	}
	if fq == nil || fq.Selector == nil {
		return nil, errors.New("Missing required key: selector")
	}
	rows := &resultSet{
		docIDs: make([]string, 0),
		revs:   make([]*revision, 0),
	}
	for docID := range d.db.docs {
		if doc, found := d.db.latestRevision(docID); found {
			rows.docIDs = append(rows.docIDs, docID)
			rows.revs = append(rows.revs, doc)
		}
	}
	rows.offset = 0
	rows.totalRows = int64(len(rows.docIDs))
	return rows, nil
}
