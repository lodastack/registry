package model

import (
	"encoding/json"
	"fmt"

	"github.com/lodastack/registry/common"
)

// Data Format: map1uuid 0 map1key1 1 map1value1 11 map1key2 1 map1value2 111 map2uuid 0 map2key1 1 map2value1 11 map2key2 1 map2value2 2

const (
	idKey      = "_id"
	Prefix int = iota
	Surffix
)

var (
	deliVal  []byte = []byte{1}
	deliProp []byte = []byte{1, 1}
	deliRes  []byte = []byte{1, 1, 1}

	uuidByte byte = 0
	deliByte byte = 1
	endByte  byte = 2
)

const (
	propertyKey int = iota
	propertyValue
)

// ResAction is the interface that resources Marshal and Unmarshal method.
type ResAction interface {
	Marshal() ([]byte, error)
	Unmarshal(raw []byte) error
}

type Resources []Resource

type Resource map[string]string

func NewResources(byteData []byte) (*Resources, error) {
	rs := &Resources{}
	resMaps := []map[string]string{}
	if err := json.Unmarshal(byteData, &resMaps); err != nil {
		return rs, fmt.Errorf("marshal bytes to map fail: %s", err.Error())
	}

	*rs = make([]Resource, len(resMaps))
	for i := 0; i < len(resMaps); i++ {
		(*rs)[i] = Resource(resMaps[i])
	}
	return rs, nil
}

func NewResourcesMaps(resMaps []map[string]string) (*Resources, error) {
	rs := &Resources{}
	*rs = make([]Resource, len(resMaps))
	for i := 0; i < len(resMaps); i++ {
		(*rs)[i] = Resource(resMaps[i])
	}
	return rs, nil
}

// Unmarshal the byte format data and stores the result
// in the value pointed to Resources.
func (rs *Resources) Unmarshal(raw []byte) error {
	*rs = make([]Resource, 0)
	startPos, endPos := 0, 0
	deliLen := 0

	for index, byt := range raw {
		switch byt {
		case deliByte:
			// Count length of deliByte.
			deliLen++
		case endByte:
			//  End of resources.
			r := Resource{}
			if err := r.Unmarshal(raw[startPos:index]); err != nil {
				return fmt.Errorf("unmarshal byte to resource fail")
			} else {
				*rs = append(*rs, r)
			}
			goto END
		default:
			if deliLen != 0 {
				switch deliLen {
				// TODO: length or another utf8 byte.
				case len(deliRes):
					endPos = index
					r := Resource{}
					if err := r.Unmarshal(raw[startPos : endPos-len(deliRes)+1]); err != nil {
						return fmt.Errorf("unmarshal byte to resource fail")
					} else {
						*rs = append(*rs, r)
					}
					startPos = index
				}
				deliLen = 0
			}
		}
	}
END:
	return nil
}

// Marshal returns the byte format data of Resources.
func (rs *Resources) Marshal() ([]byte, error) {
	raw := make([]byte, 0)
	for _, resource := range *rs {
		resourceByte, err := resource.Marshal()
		if err != nil {
			return raw, err
		}
		raw = append(raw, resourceByte...)
		raw = append(raw, deliRes...)
	}
	raw[len(raw)-len(deliRes)] = endByte
	return raw[0 : len(raw)-len(deliRes)+1], nil
}

func (rs *Resources) AppendResource(resByte []byte) error {
	r := Resource{}
	if err := r.Unmarshal(resByte); err != nil {
		return fmt.Errorf("unmarshal resource fail")
	}
	(*rs) = append((*rs), r)
	return nil
}

func (r *Resource) Unmarshal(raw []byte) error {
	tmpk, tmpv := make([]byte, 0), make([]byte, 0)
	kvFlag := propertyKey
	deliLen := 0

	for _, byt := range raw {
		switch byt {
		case uuidByte:
			// The key readed is uuid.
			(*r)[idKey] = string(tmpk)
			tmpk = make([]byte, 0)
		case deliByte:
			// Count length of deliByte.
			deliLen++
		default:
			if deliLen != 0 {
				switch deliLen {
				case len(deliVal):
					if kvFlag == propertyKey {
						kvFlag = propertyValue
					} else {
						return fmt.Errorf("unmarshal resources fail")
					}
				case len(deliProp):
					kvFlag = propertyKey
					(*r)[string(tmpk)] = string(tmpv)
					tmpk, tmpv = make([]byte, 0), make([]byte, 0)
				}
				deliLen = 0
			}
			if kvFlag == propertyKey {
				tmpk = append(tmpk, byt)
			} else {
				tmpv = append(tmpv, byt)
			}
		}
	}
	(*r)[string(tmpk)] = string(tmpv)
	return nil
}

// Marshal return byte of resource
func (r *Resource) Marshal() ([]byte, error) {
	raw := make([]byte, 0)
	uuidStr, ok := (*r)[idKey]
	uuid := []byte{}
	if ok {
		uuid = append([]byte(uuidStr), uuidByte)
	} else {
		uuid = append([]byte(common.GenUUID()), uuidByte)
	}
	raw = append(raw, uuid...)

	delete(*r, idKey)
	for k, v := range *r {
		raw = append(raw, []byte(k)...)
		raw = append(raw, deliVal...)
		raw = append(raw, []byte(v)...)
		raw = append(raw, deliProp...)
	}
	lenTotal, lenDelli := len(raw), len(deliProp)
	if lenTotal <= lenDelli {
		return nil, fmt.Errorf("marshal resource fail")
	}
	return raw[0 : lenTotal-lenDelli], nil
}

// ReadProperty return property value value of key.
func (r *Resource) ReadProperty(key string) (string, bool) {
	v, ok := (*r)[key]
	return v, ok
}
