package model

import (
	"errors"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/common"
)

// Data Format: map1uuid 0 map1key1 1 map1value1 11 map1key2 1 map1value2 111 map2uuid 0 map2key1 1 map2value1 11 map2key2 1 map2value2 2

const (
	IdKey      = "_id"
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

	ErrResFormat    error = errors.New("invalid resource format")
	ErrInvalidParam error = errors.New("invalid param")
	ErrInvalidUUID  error = errors.New("invalid uuid")
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

type ResourceList []Resource

type Resource map[string]string

func NewResourceList(resMaps []map[string]string) (*ResourceList, error) {
	rl := &ResourceList{}
	*rl = make([]Resource, len(resMaps))
	for i := 0; i < len(resMaps); i++ {
		(*rl)[i] = Resource(resMaps[i])
	}
	return rl, nil
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
type walkResourceFunc func(rByte []byte, last bool, rl *ResourceList, output []byte) ([]byte, error)

// walk the resources byte, process every resource by handler.
func (rl *ResourceList) WalkRsByte(rsByte []byte, handler walkResourceFunc) ([]byte, error) {
	*rl = make([]Resource, 0)
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
			if output, err = handler(rsByte[startPos:index], true, rl, output); err != nil {
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
					if output, err = handler(rsByte[startPos:endPos-len(deliRes)+1], false, rl, output); err != nil {
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
func (rl *ResourceList) Unmarshal(raw []byte) error {
	_, err := rl.WalkRsByte(raw, func(rByte []byte, last bool, rlWalk *ResourceList, output []byte) ([]byte, error) {
		r := Resource{}
		if err := r.Unmarshal(rByte); err != nil {
			log.Errorf("unmarshal resource fail:  %s", rByte)
			return nil, errors.New("unmarshal resources fail: " + err.Error())
		}
		*rlWalk = append(*rlWalk, r)
		return nil, nil
	})
	return err
}

// Update resource with resourceID by updateMap.
// NOTE: will not change resource ID.
func UpdateResByID(rsByte []byte, ID string, updateMap map[string]string) ([]byte, error) {
	if len(rsByte) == 0 {
		return nil, errors.New("empty resource to update")
	}
	var match bool
	return (&ResourceList{}).WalkRsByte(rsByte, func(rByte []byte, last bool, rlWalk *ResourceList, output []byte) ([]byte, error) {
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
			match = true
			for k, v := range updateMap {
				if k == IdKey {
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
			if !match {
				return nil, errors.New("not match the id when update resource")
			}
			output = append(output, rByte...)
			output = append(output, endByte)
		} else {
			output = append(output, rByte...)
			output = append(output, deliRes...)
		}
		return output, nil
	})
}

// Delete resource by resourceID..
func DeleteResource(rsByte []byte, IDs ...string) ([]byte, error) {
	return (&ResourceList{}).WalkRsByte(rsByte, func(rByte []byte, last bool, rlWalk *ResourceList, output []byte) ([]byte, error) {
		r := Resource{}
		if len(rByte) == 0 {
			return nil, errors.New("UpdateResByID fail: empty resource input")
		}
		err := r.Unmarshal(rByte)
		if err != nil {
			return nil, errors.New("UpdateResByID unmarshal resources fail: " + err.Error())
		}

		// append the resource to output if ID not matched.
		resID, _ := r.ID()
		if _, ok := common.ContainString(IDs, resID); !ok {
			rByte, err = r.Marshal()
			if err == nil {
				if len(output) != 0 {
					output = append(output, deliRes...)
				}
				output = append(output, rByte...)
			} else if err == ErrInvalidUUID {
				log.Errorf("invalid uuid in DeleteResource: %+v\n", r)
			} else {
				return nil, err
			}
		}

		if last && len(output) > 0 {
			output = append(output, endByte)
		}
		return output, nil
	})
}

// Get multi resource from rl by reousrce property.
// And Get resource by resource ID, if k is IdKey.
func (rl *ResourceList) Get(propertyK string, ValueList ...string) ([]Resource, error) {
	if len(*rl) == 0 || propertyK == "" || len(ValueList) == 0 {
		return nil, ErrInvalidParam
	}

	out := []Resource{}

	for _, r := range *rl {
		for _, v := range ValueList {
			if resV, _ := r.ReadProperty(propertyK); resV == v {
				out = append(out, r)
			}
		}
	}
	return out, nil
}

// Size returns marshed bytes size.
func (rl *ResourceList) Size() int {
	var totalSize int
	for _, resource := range *rl {
		totalSize += resource.Size()
		totalSize += len(deliRes)
	}
	return totalSize
}

// Marshal returns the byte format data of Resources.
func (rl *ResourceList) Marshal() ([]byte, error) {
	// return error when resource is empty.
	if len(*rl) == 0 {
		return nil, common.ErrEmptyResource
	}

	totalSize := rl.Size()
	raw := make([]byte, totalSize)

	var n int
	for _, resource := range *rl {
		if resource == nil {
			continue
		}
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

func (rl *ResourceList) AppendResourceByte(resByte []byte) error {
	r := Resource{}
	if err := r.Unmarshal(resByte); err != nil {
		return errors.New("unmarshal resource fail" + err.Error())
	}
	(*rl) = append((*rl), r)
	return nil
}

func (rl *ResourceList) AppendResource(r ...Resource) {
	(*rl) = append((*rl), r...)
}

func (rl *ResourceList) AppendResources(res ResourceList) {
	(*rl) = append((*rl), res...)
}

func (r *Resource) Unmarshal(raw []byte) error {
	tmpk, tmpv := make([]byte, 0), make([]byte, 0)
	kvFlag := propertyKey
	deliLen := 0

	for _, byt := range raw {
		switch byt {
		case uuidByte:
			// The key readed is uuid.
			(*r)[IdKey] = string(tmpk)
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
				case len(deliRes):
					return errors.New("invalid deliLen in resource Unmarshal")
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
	// Process the last property.
	if len(tmpk) > 0 {
		(*r)[string(tmpk)] = string(tmpv)
	}
	return nil
}

// Size returns marshed bytes size.
func (r *Resource) Size() int {
	// string UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// and UUID flag: uuidByte
	totalSize := 36 + 1
	for k, v := range *r {
		if k == IdKey {
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
	if len(UUID) != 36 {
		return nil, ErrInvalidUUID
	}
	n += copy(raw[n:], []byte(UUID))
	raw[n] = uuidByte
	n += 1

	for k, v := range *r {
		if k == IdKey {
			continue
		}
		n += copy(raw[n:], []byte(k))
		n += copy(raw[n:], deliVal)

		// If value is empty, take a nil byte.
		// Avoid deliVal and deliProp combine into deliRes.
		if len(v) == 0 {
			raw[n] = nilByte
			n += 1
		} else {
			n += copy(raw[n:], []byte(v))
		}
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

// RemoveProperty set the k-v to resource.
func (r *Resource) RemoveProperty(k string) {
	delete((*r), k)
}

// InitID create ID for the resource if not have, and return ID.
func (r *Resource) InitID() string {
	id, _ := r.ID()
	if id == "" {
		return r.NewID()
	}
	return id
}

func (r *Resource) NewID() string {
	id := common.GenUUID()
	r.SetProperty(IdKey, id)
	return id
}

func (r *Resource) ID() (string, bool) {
	return r.ReadProperty(IdKey)
}

// ResourcesAppendByte append the resource to resources.
func AppendResources(rsByte []byte, rs ...Resource) ([]byte, error) {
	for i := range rs {
		rs[i].InitID()
		if len(rs[i]) <= 1 {
			return nil, errors.New("not allow append resource only have id")
		}
	}

	rl := ResourceList{}
	if err := rl.Unmarshal(rsByte); err != nil {
		return nil, err
	}
	rl.AppendResource(rs...)
	return rl.Marshal()
}
