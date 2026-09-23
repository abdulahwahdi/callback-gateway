package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap is a generic string-keyed map persisted as a jsonb column.
// Used for storing request headers and query parameters.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	b, ok := asBytes(src)
	if !ok {
		return errors.New("domain: JSONMap.Scan: unsupported source type")
	}
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(b, m)
}

// JSONRaw stores an arbitrary raw JSON payload (the exact webhook body sent
// by the payment gateway) as jsonb, without re-shaping it.
type JSONRaw json.RawMessage

func (r JSONRaw) Value() (driver.Value, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return []byte(r), nil
}

func (r *JSONRaw) Scan(src any) error {
	if src == nil {
		*r = nil
		return nil
	}
	b, ok := asBytes(src)
	if !ok {
		return errors.New("domain: JSONRaw.Scan: unsupported source type")
	}
	*r = JSONRaw(append([]byte(nil), b...))
	return nil
}

func (r JSONRaw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *JSONRaw) UnmarshalJSON(data []byte) error {
	*r = append((*r)[0:0], data...)
	return nil
}

// StringArray persists a []string as a jsonb array (used for the list of
// Kafka topics a webhook event was published to).
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = StringArray{}
		return nil
	}
	b, ok := asBytes(src)
	if !ok {
		return errors.New("domain: StringArray.Scan: unsupported source type")
	}
	if len(b) == 0 {
		*a = StringArray{}
		return nil
	}
	return json.Unmarshal(b, a)
}

func asBytes(src any) ([]byte, bool) {
	switch v := src.(type) {
	case []byte:
		return v, true
	case string:
		return []byte(v), true
	default:
		return nil, false
	}
}
