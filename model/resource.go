package model

import (
	"encoding/json"
	"errors"

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
	nilByte  byte = 3
)

var (
	ErrResMarshal error = errors.New("marshal resources fail")
	ErrEmptyRes   error = errors.New("empty resources")
	ErrResFormat  error = errors.New("invalid resource fromat")
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
		return rs, errors.New("marshal bytes to map fail: " + err.Error())
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

func NewResource(resMap map[string]string) Resource {
	addRes := Resource(resMap)
	return addRes
}

// walkResourceFunc is the type of the function for each resource byte  visited by WalkRsByte.
// The rByte argument is the byte of a resource.
// The rs argument is the pointer to the method caller.
//
// If an error was returned, processing stops.
type walkResourceFunc func(rByte []byte, last bool, rs *Resources, output []byte) ([]byte, error)

// walk the resources byte, process every resource by handler.
func (rs *Resources) WalkRsByte(rsByte []byte, handler walkResourceFunc) ([]byte, error) {
	*rs = make([]Resource, 0)
	startPos, endPos := 0, 0
	deliLen := 0
	output := make([]byte, 0)
	var err error

	for index, byt := range rsByte {
		switch byt {
		case deliByte:
			// Count length of deliByte.
			deliLen++
		case endByte:
			//  End of resources.
			if output, err = handler(rsByte[startPos:index], true, rs, output); err != nil {
				return nil, errors.New("process resource fail: " + err.Error())
			}
			goto END
		case nilByte:
			// Read value is done if read a nilByte.
			fallthrough
		default:
			if deliLen != 0 {
				switch deliLen {
				// TODO: length or another utf8 byte.
				case len(deliRes):
					endPos = index
					if output, err = handler(rsByte[startPos:endPos-len(deliRes)+1], false, rs, output); err != nil {
						return nil, errors.New("process resource fail: " + err.Error())
					}
					startPos = index
				}
				deliLen = 0
			}
		}
	}
END:
	return output, nil
}

// Unmarshal the byte to the method caller rs.
func (rs *Resources) Unmarshal(raw []byte) error {
	_, err := rs.WalkRsByte(raw, func(rByte []byte, last bool, rs *Resources, output []byte) ([]byte, error) {
		r := Resource{}
		if err := r.Unmarshal(rByte); err != nil {
			return nil, errors.New("unmarshal resources fail: " + err.Error())
		}
		*rs = append(*rs, r)
		return nil, nil
	})
	return err
}

// Update resource with resourceID by updateMap.
// NOTE: will not change resource ID.
func UpdateResByID(rsByte []byte, ID string, updateMap map[string]string) ([]byte, error) {
	return (&Resources{}).WalkRsByte(rsByte, func(rByte []byte, last bool, rs *Resources, output []byte) ([]byte, error) {
		r := Resource{}
		if len(rByte) == 0 {
			return nil, errors.New("UpdateResByID fail: empty resource input")
		}
		err := r.Unmarshal(rByte)
		if err != nil {
			return nil, errors.New("UpdateResByID unmarshal resources fail: " + err.Error())
		}

		// update the resource if resource ID match with expect.
		if resID, _ := r.ID(); resID == ID {
			for k, v := range updateMap {
				if k == idKey {
					continue
				}
				r.SetProperty(k, v)
			}
		}

		rByte, err = r.Marshal()
		if err != nil {
			return nil, err
		}
		if last {
			output = append(output, rByte...)
			output = append(output, endByte)
		} else {
			output = append(output, rByte...)
			output = append(output, deliRes...)
		}
		return output, nil
	})
}

// Size returns marshed bytes size.
func (rs *Resources) Size() int {
	var totalSize int
	for _, resource := range *rs {
		totalSize += resource.Size()
		totalSize += len(deliRes)
	}
	return totalSize
}

// Marshal returns the byte format data of Resources.
func (rs *Resources) Marshal() ([]byte, error) {
	// return error when resource is empty.
	if len(*rs) == 0 {
		return nil, ErrEmptyRes
	}

	totalSize := rs.Size()
	raw := make([]byte, totalSize)

	var n int
	for _, resource := range *rs {
		resourceByte, err := resource.Marshal()
		if err != nil {
			return raw, err
		}
		n += copy(raw[n:], resourceByte)
		n += copy(raw[n:], deliRes)
	}
	raw[n-len(deliRes)] = endByte
	return raw[0 : n-len(deliRes)+1], nil
}

func (rs *Resources) AppendResourceByte(resByte []byte) error {
	r := Resource{}
	if err := r.Unmarshal(resByte); err != nil {
		return errors.New("unmarshal resource fail")
	}
	(*rs) = append((*rs), r)
	return nil
}

func (rs *Resources) AppendResource(r Resource) {
	(*rs) = append((*rs), r)
}

func (rs *Resources) AppendResources(res Resources) {
	(*rs) = append((*rs), res...)
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
		case nilByte:
			fallthrough
		default:
			if deliLen != 0 {
				switch deliLen {
				case len(deliVal):
					if kvFlag == propertyKey {
						kvFlag = propertyValue
					} else {
						return errors.New("unmarshal resource fail")
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
			} else if byt != nilByte {
				tmpv = append(tmpv, byt)
			}
		}
	}
	(*r)[string(tmpk)] = string(tmpv)
	return nil
}

// Size returns marshed bytes size.
func (r *Resource) Size() int {
	// string UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// and UUID flag: uuidByte
	totalSize := 36 + 1
	for k, v := range *r {
		if k == idKey {
			continue
		}
		totalSize += len(k)
		totalSize += len(deliVal)

		// If value is empty, take a nil byte.
		// Avoid deliVal and deliProp combine into deliRes.
		if len(v) == 0 {
			totalSize += 1
		}
		totalSize += len(v)
		totalSize += len(deliProp)
	}
	return totalSize
}

// Marshal will create UUID if the resource have no ID.
// Return the resource []byte/ID.
func (r *Resource) Marshal() ([]byte, error) {
	totalSize := r.Size()
	raw := make([]byte, totalSize)
	var n int

	UUID := r.InitID()
	n += copy(raw[n:], []byte(UUID))
	raw[n] = uuidByte
	n += 1

	for k, v := range *r {
		if k == idKey {
			continue
		}
		n += copy(raw[n:], []byte(k))
		n += copy(raw[n:], deliVal)

		// If value is empty, take a nil byte.
		// Avoid deliVal and deliProp combine into deliRes.
		if len(v) == 0 {
			raw[n] = nilByte
			n += 1
		}

		n += copy(raw[n:], []byte(v))
		n += copy(raw[n:], deliProp)
	}
	return raw[0 : n-len(deliProp)], nil
}

// ReadProperty return property value value of key.
func (r *Resource) ReadProperty(key string) (string, bool) {
	v, ok := (*r)[key]
	return v, ok
}

// SetProperty set the k-v to resource.
func (r *Resource) SetProperty(k, v string) {
	(*r)[k] = v
}

// InitID create ID for the resource if not have, and return ID.
func (r *Resource) InitID() string {
	if id, _ := r.ID(); id == "" {
		(*r)[idKey] = common.GenUUID()
	}
	return (*r)[idKey]
}

func (r *Resource) ID() (string, bool) {
	return r.ReadProperty(idKey)
}

func delEndByte(ori []byte) ([]byte, error) {
	oriLen := len(ori)
	if ori[oriLen-1] != endByte {
		return nil, ErrResFormat
	}
	return ori[:oriLen-1], nil
}

// ResourcesAppendByte append the resource to resources.
func AppendResources(rsByte []byte, resource Resource) ([]byte, string, error) {
	UUID := resource.InitID()

	// If append res to nil, new resources.
	if len(rsByte) == 0 {
		rs := Resources{}
		rs.AppendResource(resource)
		rsByte, err := rs.Marshal()
		return rsByte, UUID, err
	}

	// rm the endByte of resource
	resNoEnd, err := delEndByte(rsByte)
	if err != nil {
		return nil, "", err
	}
	// append deliRes/new resource/endByte
	addLen := len(deliRes) + resource.Size() + 1
	addByte := make([]byte, addLen)
	resByte, err := resource.Marshal()
	if err != nil {
		return nil, "", ErrResMarshal
	}
	n := copy(addByte, deliRes)
	n += copy(addByte[n:], resByte)
	addByte[n] = endByte

	return append(resNoEnd, addByte[:n+1]...), UUID, nil
}
