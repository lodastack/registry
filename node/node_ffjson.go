// DO NOT EDIT!
// Code generated by ffjson <https://github.com/pquerna/ffjson>
// source: node.go
// DO NOT EDIT!

package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

func (mj *Node) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if mj == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := mj.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (mj *Node) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if mj == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"Children":`)
	if mj.Children != nil {
		buf.WriteString(`[`)
		for i, v := range mj.Children {
			if i != 0 {
				buf.WriteString(`,`)
			}

			{

				if v == nil {
					buf.WriteString("null")
					return nil
				}

				err = v.MarshalJSONBuf(buf)
				if err != nil {
					return err
				}

			}
		}
		buf.WriteString(`]`)
	} else {
		buf.WriteString(`null`)
	}
	buf.WriteString(`,"ID":`)
	fflib.WriteJsonString(buf, string(mj.ID))
	buf.WriteString(`,"Name":`)
	fflib.WriteJsonString(buf, string(mj.Name))
	buf.WriteString(`,"Type":`)
	fflib.FormatBits2(buf, uint64(mj.Type), 10, mj.Type < 0)
	buf.WriteString(`,"MachineReg":`)
	fflib.WriteJsonString(buf, string(mj.MachineReg))
	buf.WriteByte('}')
	return nil
}

const (
	ffj_t_Nodebase = iota
	ffj_t_Nodeno_such_key

	ffj_t_Node_Children

	ffj_t_Node_ID

	ffj_t_Node_Name

	ffj_t_Node_Type

	ffj_t_Node_MachineReg
)

var ffj_key_Node_Children = []byte("Children")

var ffj_key_Node_ID = []byte("ID")

var ffj_key_Node_Name = []byte("Name")

var ffj_key_Node_Type = []byte("Type")

var ffj_key_Node_MachineReg = []byte("MachineReg")

func (uj *Node) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *Node) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_Nodebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_Nodeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'C':

					if bytes.Equal(ffj_key_Node_Children, kn) {
						currentKey = ffj_t_Node_Children
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'I':

					if bytes.Equal(ffj_key_Node_ID, kn) {
						currentKey = ffj_t_Node_ID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'M':

					if bytes.Equal(ffj_key_Node_MachineReg, kn) {
						currentKey = ffj_t_Node_MachineReg
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'N':

					if bytes.Equal(ffj_key_Node_Name, kn) {
						currentKey = ffj_t_Node_Name
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'T':

					if bytes.Equal(ffj_key_Node_Type, kn) {
						currentKey = ffj_t_Node_Type
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffj_key_Node_MachineReg, kn) {
					currentKey = ffj_t_Node_MachineReg
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Node_Type, kn) {
					currentKey = ffj_t_Node_Type
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Node_Name, kn) {
					currentKey = ffj_t_Node_Name
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Node_ID, kn) {
					currentKey = ffj_t_Node_ID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Node_Children, kn) {
					currentKey = ffj_t_Node_Children
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_Nodeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_Node_Children:
					goto handle_Children

				case ffj_t_Node_ID:
					goto handle_ID

				case ffj_t_Node_Name:
					goto handle_Name

				case ffj_t_Node_Type:
					goto handle_Type

				case ffj_t_Node_MachineReg:
					goto handle_MachineReg

				case ffj_t_Nodeno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Children:

	/* handler: uj.Children type=[]*node.Node kind=slice quoted=false*/

	{

		{
			if tok != fflib.FFTok_left_brace && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for ", tok))
			}
		}

		if tok == fflib.FFTok_null {
			uj.Children = nil
		} else {

			uj.Children = make([]*Node, 0)

			wantVal := true

			for {

				var v *Node

				tok = fs.Scan()
				if tok == fflib.FFTok_error {
					goto tokerror
				}
				if tok == fflib.FFTok_right_brace {
					break
				}

				if tok == fflib.FFTok_comma {
					if wantVal == true {
						// TODO(pquerna): this isn't an ideal error message, this handles
						// things like [,,,] as an array value.
						return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
					}
					continue
				} else {
					wantVal = true
				}

				/* handler: v type=*node.Node kind=ptr quoted=false*/

				{
					if tok == fflib.FFTok_null {

						v = nil

						state = fflib.FFParse_after_value
						goto mainparse
					}

					if v == nil {
						v = new(Node)
					}

					err = v.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
					if err != nil {
						return err
					}
					state = fflib.FFParse_after_value
				}

				uj.Children = append(uj.Children, v)
				wantVal = false
			}
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ID:

	/* handler: uj.ID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.ID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Name:

	/* handler: uj.Name type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Name = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Type:

	/* handler: uj.Type type=int kind=int quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.Type = int(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_MachineReg:

	/* handler: uj.MachineReg type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.MachineReg = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}

func (mj *NodeProperty) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if mj == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := mj.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (mj *NodeProperty) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if mj == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"ID":`)
	fflib.WriteJsonString(buf, string(mj.ID))
	buf.WriteString(`,"Name":`)
	fflib.WriteJsonString(buf, string(mj.Name))
	buf.WriteString(`,"Type":`)
	fflib.FormatBits2(buf, uint64(mj.Type), 10, mj.Type < 0)
	buf.WriteString(`,"MachineReg":`)
	fflib.WriteJsonString(buf, string(mj.MachineReg))
	buf.WriteByte('}')
	return nil
}

const (
	ffj_t_NodePropertybase = iota
	ffj_t_NodePropertyno_such_key

	ffj_t_NodeProperty_ID

	ffj_t_NodeProperty_Name

	ffj_t_NodeProperty_Type

	ffj_t_NodeProperty_MachineReg
)

var ffj_key_NodeProperty_ID = []byte("ID")

var ffj_key_NodeProperty_Name = []byte("Name")

var ffj_key_NodeProperty_Type = []byte("Type")

var ffj_key_NodeProperty_MachineReg = []byte("MachineReg")

func (uj *NodeProperty) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *NodeProperty) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_NodePropertybase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_NodePropertyno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'I':

					if bytes.Equal(ffj_key_NodeProperty_ID, kn) {
						currentKey = ffj_t_NodeProperty_ID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'M':

					if bytes.Equal(ffj_key_NodeProperty_MachineReg, kn) {
						currentKey = ffj_t_NodeProperty_MachineReg
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'N':

					if bytes.Equal(ffj_key_NodeProperty_Name, kn) {
						currentKey = ffj_t_NodeProperty_Name
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'T':

					if bytes.Equal(ffj_key_NodeProperty_Type, kn) {
						currentKey = ffj_t_NodeProperty_Type
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffj_key_NodeProperty_MachineReg, kn) {
					currentKey = ffj_t_NodeProperty_MachineReg
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_NodeProperty_Type, kn) {
					currentKey = ffj_t_NodeProperty_Type
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_NodeProperty_Name, kn) {
					currentKey = ffj_t_NodeProperty_Name
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_NodeProperty_ID, kn) {
					currentKey = ffj_t_NodeProperty_ID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_NodePropertyno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_NodeProperty_ID:
					goto handle_ID

				case ffj_t_NodeProperty_Name:
					goto handle_Name

				case ffj_t_NodeProperty_Type:
					goto handle_Type

				case ffj_t_NodeProperty_MachineReg:
					goto handle_MachineReg

				case ffj_t_NodePropertyno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_ID:

	/* handler: uj.ID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.ID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Name:

	/* handler: uj.Name type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Name = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Type:

	/* handler: uj.Type type=int kind=int quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.Type = int(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_MachineReg:

	/* handler: uj.MachineReg type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.MachineReg = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}

func (mj *Tree) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if mj == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := mj.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (mj *Tree) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if mj == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	if mj.Nodes != nil {
		buf.WriteString(`{"Nodes":`)

		{

			err = mj.Nodes.MarshalJSONBuf(buf)
			if err != nil {
				return err
			}

		}
	} else {
		buf.WriteString(`{"Nodes":null`)
	}
	buf.WriteString(`,"Cluster":`)
	/* Interface types must use runtime reflection. type=node.Cluster kind=interface */
	err = buf.Encode(mj.Cluster)
	if err != nil {
		return err
	}
	if mj.RWsync != nil {
		/* Struct fall back. type=sync.RWMutex kind=struct */
		buf.WriteString(`,"RWsync":`)
		err = buf.Encode(mj.RWsync)
		if err != nil {
			return err
		}
	} else {
		buf.WriteString(`,"RWsync":null`)
	}
	buf.WriteByte('}')
	return nil
}

const (
	ffj_t_Treebase = iota
	ffj_t_Treeno_such_key

	ffj_t_Tree_Nodes

	ffj_t_Tree_Cluster

	ffj_t_Tree_RWsync
)

var ffj_key_Tree_Nodes = []byte("Nodes")

var ffj_key_Tree_Cluster = []byte("Cluster")

var ffj_key_Tree_RWsync = []byte("RWsync")

func (uj *Tree) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *Tree) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_Treebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_Treeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'C':

					if bytes.Equal(ffj_key_Tree_Cluster, kn) {
						currentKey = ffj_t_Tree_Cluster
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'N':

					if bytes.Equal(ffj_key_Tree_Nodes, kn) {
						currentKey = ffj_t_Tree_Nodes
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'R':

					if bytes.Equal(ffj_key_Tree_RWsync, kn) {
						currentKey = ffj_t_Tree_RWsync
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffj_key_Tree_RWsync, kn) {
					currentKey = ffj_t_Tree_RWsync
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Tree_Cluster, kn) {
					currentKey = ffj_t_Tree_Cluster
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Tree_Nodes, kn) {
					currentKey = ffj_t_Tree_Nodes
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_Treeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_Tree_Nodes:
					goto handle_Nodes

				case ffj_t_Tree_Cluster:
					goto handle_Cluster

				case ffj_t_Tree_RWsync:
					goto handle_RWsync

				case ffj_t_Treeno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Nodes:

	/* handler: uj.Nodes type=node.Node kind=struct quoted=false*/

	{
		if tok == fflib.FFTok_null {

			uj.Nodes = nil

			state = fflib.FFParse_after_value
			goto mainparse
		}

		if uj.Nodes == nil {
			uj.Nodes = new(Node)
		}

		err = uj.Nodes.UnmarshalJSONFFLexer(fs, fflib.FFParse_want_key)
		if err != nil {
			return err
		}
		state = fflib.FFParse_after_value
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Cluster:

	/* handler: uj.Cluster type=node.Cluster kind=interface quoted=false*/

	{
		/* Falling back. type=node.Cluster kind=interface */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &uj.Cluster)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_RWsync:

	/* handler: uj.RWsync type=sync.RWMutex kind=struct quoted=false*/

	{
		/* Falling back. type=sync.RWMutex kind=struct */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &uj.RWsync)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}

func (mj *nodeCache) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if mj == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := mj.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (mj *nodeCache) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if mj == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	if mj.Cache != nil {
		if *mj.Cache == nil {
			buf.WriteString(`{"Cache":null`)
		} else {
			buf.WriteString(`{"Cache":{ `)
			for key, value := range *mj.Cache {
				fflib.WriteJsonString(buf, key)
				buf.WriteString(`:`)
				fflib.WriteJsonString(buf, string(value))
				buf.WriteByte(',')
			}
			buf.Rewind(1)
			buf.WriteByte('}')
		}
	} else {
		buf.WriteString(`{"Cache":null`)
	}
	if mj.RWsync != nil {
		/* Struct fall back. type=sync.RWMutex kind=struct */
		buf.WriteString(`,"RWsync":`)
		err = buf.Encode(mj.RWsync)
		if err != nil {
			return err
		}
	} else {
		buf.WriteString(`,"RWsync":null`)
	}
	buf.WriteByte('}')
	return nil
}

const (
	ffj_t_nodeCachebase = iota
	ffj_t_nodeCacheno_such_key

	ffj_t_nodeCache_Cache

	ffj_t_nodeCache_RWsync
)

var ffj_key_nodeCache_Cache = []byte("Cache")

var ffj_key_nodeCache_RWsync = []byte("RWsync")

func (uj *nodeCache) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *nodeCache) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_nodeCachebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_nodeCacheno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'C':

					if bytes.Equal(ffj_key_nodeCache_Cache, kn) {
						currentKey = ffj_t_nodeCache_Cache
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'R':

					if bytes.Equal(ffj_key_nodeCache_RWsync, kn) {
						currentKey = ffj_t_nodeCache_RWsync
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffj_key_nodeCache_RWsync, kn) {
					currentKey = ffj_t_nodeCache_RWsync
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_nodeCache_Cache, kn) {
					currentKey = ffj_t_nodeCache_Cache
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_nodeCacheno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_nodeCache_Cache:
					goto handle_Cache

				case ffj_t_nodeCache_RWsync:
					goto handle_RWsync

				case ffj_t_nodeCacheno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Cache:

	/* handler: uj.Cache type=map[string]string kind=map quoted=false*/

	{

		{
			if tok != fflib.FFTok_left_bracket && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for ", tok))
			}
		}

		if tok == fflib.FFTok_null {
			uj.Cache = nil
		} else {

			var tval = make(map[string]string, 0)

			wantVal := true

			for {

				var k string

				var v string

				tok = fs.Scan()
				if tok == fflib.FFTok_error {
					goto tokerror
				}
				if tok == fflib.FFTok_right_bracket {
					break
				}

				if tok == fflib.FFTok_comma {
					if wantVal == true {
						// TODO(pquerna): this isn't an ideal error message, this handles
						// things like [,,,] as an array value.
						return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
					}
					continue
				} else {
					wantVal = true
				}

				/* handler: k type=string kind=string quoted=false*/

				{

					{
						if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
							return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
						}
					}

					if tok == fflib.FFTok_null {

					} else {

						outBuf := fs.Output.Bytes()

						k = string(string(outBuf))

					}
				}

				// Expect ':' after key
				tok = fs.Scan()
				if tok != fflib.FFTok_colon {
					return fs.WrapErr(fmt.Errorf("wanted colon token, but got token: %v", tok))
				}

				tok = fs.Scan()
				/* handler: v type=string kind=string quoted=false*/

				{

					{
						if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
							return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
						}
					}

					if tok == fflib.FFTok_null {

					} else {

						outBuf := fs.Output.Bytes()

						v = string(string(outBuf))

					}
				}

				tval[k] = v

				wantVal = false
			}

			uj.Cache = &tval

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_RWsync:

	/* handler: uj.RWsync type=sync.RWMutex kind=struct quoted=false*/

	{
		/* Falling back. type=sync.RWMutex kind=struct */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &uj.RWsync)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}
