package dedup

import (
	"encoding/json"
	"io"
	"os"
	"sort"
)

type DeDup[T any] struct {
	record    map[string]bool
	newRecord map[string]bool

	hashFunc func(elem T) string
}

func NewDeDup[T any](filepath string, hashFunc func(elem T) string) (*DeDup[T], error) {
	record := make(map[string]bool)
	_, err := os.Stat(filepath)
	if !os.IsNotExist(err) {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		contents, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		var list []string
		if err = json.Unmarshal(contents, &list); err != nil {
			return nil, err
		}
		for _, digest := range list {
			record[digest] = true
		}
	}

	newRecord := make(map[string]bool)
	return &DeDup[T]{record: record, newRecord: newRecord, hashFunc: hashFunc}, nil
}

func (d *DeDup[T]) FilterMany(elements []T) []T {
	var newElements []T
	for _, elem := range elements {
		digest := d.hashFunc(elem)
		if d.record[digest] {
			continue
		}

		d.newRecord[digest] = true
		newElements = append(newElements, elem)
	}
	return newElements
}

// MergeAndDump merge record map and new record map
// finally dumps
func (d *DeDup[T]) MergeAndDump(filepath string) error {
	// merge two records
	var list []string
	for k := range d.record {
		list = append(list, k)
	}
	for k := range d.newRecord {
		list = append(list, k)
	}
	sort.Strings(list)

	// dump data
	content, err := json.Marshal(list)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}
