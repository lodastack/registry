package model

import (
	"fmt"
)

type HandleFunc func(raw []byte) (Resources, error)

type ResourceSearch struct {
	Id    string // key of resource property
	Key   string // search string
	Value []byte // match prefix or Surffix

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
	kvFlag, deliLen := kPosi, 0
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
					kvFlag = kPosi
				}
				deliLen = 0
			}
			if kvFlag == kPosi {
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
	kvFlag, deliLen, matchPos := kPosi, 0, 0
	startPos, endPos := 0, 0
	matchFlag := false  // flag of the k-v in the resource(map) is matched
	matchValue := false // flag of key is matched

	for ind, byt := range raw {
		switch byt {
		case uuidByte:
			tmpk = make([]byte, 0)
		case deliByte:
			// Count length of deliByte.
			deliLen++
		case endByte:
			//  End of resources.
			if matchFlag { // TODO: end bytes with a end_byte can eliminate this process.
				if err := matchReses.AppendResource(raw[startPos:]); err != nil {
					return matchReses, fmt.Errorf("unmarshal resource fail")
				}
			}
			goto END
		default:
			if deliLen != 0 {
				switch deliLen {
				case len(deliVal):
					if kvFlag == kPosi {
						kvFlag = vPosi
					} else {
						return matchReses, fmt.Errorf("unmarshal resources fail")
					}
					if len(s.Key) == 0 || s.Key == string(tmpk) {
						// If Key is not set or is matched, value of the key should be checked.
						matchValue = true
					}
				case len(deliProp):
					matchValue = false
					kvFlag, matchPos = kPosi, 0
					tmpk = make([]byte, 0)
				case len(deliRes):
					if matchFlag {
						// The map match the search, append this resource to resource.
						endPos = ind - 3
						if err := matchReses.AppendResource(raw[startPos : endPos+1]); err != nil {
							return matchReses, fmt.Errorf("unmarshal resource fail")
						}
					}
					matchFlag, matchValue = false, false
					startPos = ind
					kvFlag, matchPos = kPosi, 0
					tmpk = make([]byte, 0)
				}
				deliLen = 0
			}

			if kvFlag == kPosi {
				tmpk = append(tmpk, byt)
			} else {
				if matchValue && s.Value[matchPos] == byt {
					matchPos++
					if matchPos == len(s.Value) {
						// if the s.Value is complete matched, the map is matched.
						matchFlag = true
					}
				} else {
					matchPos = 0
				}
			}
		}
	}
END:
	return matchReses, nil
}
