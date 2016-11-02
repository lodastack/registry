package model

import (
	"fmt"
)

type HandleFunc func(raw []byte) (Resources, error)

type ResourceSearch struct {
	Id    string // key of resource property
	Key   string // search string
	Value []byte // match prefix or Surffix
	Fuzzy bool

	Process HandleFunc
}

func (s *ResourceSearch) Init() error {
	lenId := len(s.Id)
	lenValue := len(s.Value)

	if lenValue == 0 && lenId != 0 {
		s.Process = s.IdSearch
	} else if lenValue != 0 {
		s.Process = s.ValueSearch
	} else {
		return fmt.Errorf("none value to search")
	}
	return nil
}

func (s *ResourceSearch) IdSearch(raw []byte) (Resources, error) {
	matchReses := Resources{}
	kvFlag, deliLen := propertyKey, 0
	startPos, endPos := 0, 0
	matchFlag := false
	tmpk := make([]byte, 0)

	for index, byt := range raw {
		switch byt {
		case uuidByte:
			if s.Id == string(tmpk) {
				matchFlag = true
			}
			tmpk = make([]byte, 0)
		case deliByte:
			// Count length of deliByte.
			deliLen++
		case endByte:
			//  End of resources.
			if matchFlag {
				if err := matchReses.AppendResource(raw[startPos:]); err != nil {
					return matchReses, fmt.Errorf("unmarshal resource fail")
				}
			}
			goto END
		default:
			if deliLen != 0 {
				switch deliLen {
				case len(deliRes):
					if matchFlag {
						endPos = index - 3
						if err := matchReses.AppendResource(raw[startPos : endPos+1]); err != nil {
							return matchReses, fmt.Errorf("unmarshal resource fail")
						}
					}
					tmpk = make([]byte, 0)
					matchFlag = false
					startPos = index
					kvFlag = propertyKey
				}
				deliLen = 0
			}
			if kvFlag == propertyKey {
				tmpk = append(tmpk, byt)
			}
		}
	}
END:
	return matchReses, nil
}

func (s *ResourceSearch) ValueSearch(raw []byte) (Resources, error) {
	matchReses := Resources{}
	tmpk := make([]byte, 0)
	kvFlag, deliLen := propertyKey, 0 // kvFlag is flag of byte readed is k or v.
	startPos, vPos := 0, 0            // startPos is position where resource start. vPos is position where value start.
	matchFlag := false                // flag of the k-v in the resource(map) is matched
	matchValue := false               // flag of key is matched

	//  Read the end of one resource, process the last value and push matched resoutce to result.
	processResource := func(lastValutStartPos, resStartPos, end int) error {
		// Search the last value if the resource is not match and the last value is need to search.
		if matchValue && !matchFlag {
			matchFlag = search(raw[lastValutStartPos:end], s.Value, s.Fuzzy)
		}
		// If the resource is matched, append it to result.
		if matchFlag {
			if err := matchReses.AppendResource(raw[resStartPos:end]); err != nil {
				return fmt.Errorf("unmarshal resource fail")
			}
		}
		return nil
	}

	for index := range raw {
		switch raw[index] {
		// Read the deli between uuid and resource data.
		case uuidByte:
			tmpk = make([]byte, 0)
		// Read a deli, the length of deli will indicate the type of the deli.
		case deliByte:
			// Count length of deliByte.
			deliLen++
		// Read the end of resources.
		case endByte:
			if err := processResource(vPos, startPos, index); err != nil {
				return nil, fmt.Errorf("unmarshal resource fail")
			}
			goto END
		default:
			if deliLen != 0 {
				switch deliLen {
				// Begin to read key when get a deli between k and v.
				case len(deliVal):
					if kvFlag == propertyKey {
						kvFlag = propertyValue
						vPos = index
					} else {
						// If get one deli when read value, return error.
						return nil, fmt.Errorf("unmarshal resources fail")
					}
					// If Key is not set or matched, value of the key should be checked.
					if len(s.Key) == 0 || s.Key == string(tmpk) {
						matchValue = true
					}
				// Read value completely when get a deliProp which a deli between kv pairs.
				case len(deliProp):
					kvFlag = propertyKey
					tmpk = make([]byte, 0)
					// If the value neet to search.
					if matchValue {
						matchValue = false
						matchFlag = search(raw[vPos:index-deliLen], s.Value, s.Fuzzy)
					}
				// Read resource completely when get a deliRes
				case len(deliRes):
					if err := processResource(vPos, startPos, index-deliLen); err != nil {
						return nil, fmt.Errorf("unmarshal resource fail")
					}
					matchFlag, matchValue = false, false
					startPos = index
					kvFlag = propertyKey
					tmpk = make([]byte, 0)
				}
				deliLen = 0
			}

			if kvFlag == propertyKey {
				tmpk = append(tmpk, raw[index])
			}
		}
	}
END:
	return matchReses, nil
}

func search(ori, dest []byte, fuzzy bool) bool {
	if !fuzzy {
		return string(ori) == string(dest)
	}

	indexMatch := 0
	lenDest := len(dest)
	for i := range ori {
		if ori[i] == dest[indexMatch] {
			if indexMatch+1 == lenDest {
				return true
			}
			indexMatch++
		} else {
			if indexMatch != 0 {
				indexMatch = 0
			}
		}
	}
	return false
}
