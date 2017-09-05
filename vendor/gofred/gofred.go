package gofred

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Response has all the items for showing on alfred
type Response struct {
	VarMap map[string]string `json:"variables,omitempty"`
	Items  []Item            `json:"items,omitempty"`
}

// NewResponse returns a instance Response
func NewResponse() *Response {
	resp := &Response{}
	resp.VarMap = make(map[string]string)
	return resp
}

// NewItem create a new item with basic information
func NewItem(title, subtitle, autocomplete string) Item {
	return Item{
		Title:        title,
		Autocomplete: autocomplete,
		SubInfo: SubInfo{
			Subtitle: subtitle,
			Valid:    false,
		},
	}
}

// NewItemOnce returns a item to be shown
func NewItemOnce(title, subtitle, iconType, iconPath, arg, autocomplete, uid, itemType string, valid bool, mods Modifiers) Item {
	return Item{
		Title: title,
		Icon: IconInfo{
			Type: iconType,
			Path: iconPath,
		},
		Autocomplete: autocomplete,
		UID:          uid,
		Type:         itemType,
		SubInfo: SubInfo{
			Subtitle: subtitle,
			Valid:    valid,
			Arg:      arg,
		},
		Mods: mods,
	}
}

// AddVariable add a alfred environment value to pass
func (r *Response) AddVariable(key, value string) {
	r.VarMap[key] = value
}

// AddItems add a item to response
func (r *Response) AddItems(items ...Item) {
	r.Items = append(r.Items, items...)
}

// AddMatchedItems add a item if the title matches with given command
func (r *Response) AddMatchedItems(str string, items ...Item) {
	for _, item := range items {
		if len(str) == 0 || strings.Contains(item.Title, str) {
			r.AddItems(item)
		}
	}
}

func (r *Response) String() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return string(bytes)
}

// IsEmpty returns if it's empty or not
func (r *Response) IsEmpty() bool {
	return len(r.Items) == 0
}
