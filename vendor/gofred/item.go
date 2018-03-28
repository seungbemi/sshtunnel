package gofred

type modKey string

const (
	modKeyCommand = modKey("cmd")
	modKeyOption  = modKey("alt")
	modKeyControl = modKey("ctrl")
)

// Variable key, value to set var map
type Variable struct {
	key   string
	value string
}

// IconInfo includes item's icon type and path
type IconInfo struct {
	Type string `json:"type,omitempty"` // Optional
	Path string `json:"path"`
}

// SubInfo include optional information
type SubInfo struct {
	Subtitle string            `json:"subtitle,omitempty"` // Optional
	VarMap   map[string]string `json:"variables,omitempty"`
	Arg      string            `json:"arg,omitempty"` // Recommended
	Valid    bool              `json:"valid"`         // Default = true
}

// Modifiers includes subinfo for modifiers
// SubInfo can be changed by pressing a followed modifier key
type Modifiers struct {
	OptionKey  SubInfo `json:"alt,omitempty"`
	CommandKey SubInfo `json:"cmd,omitempty"`
	CtrlKey    SubInfo `json:"ctrl,omitempty"`
}

// Item that will be shown as a result
type Item struct {
	SubInfo      `json:",inline"`
	Title        string    `json:"title"` // Essential
	Icon         IconInfo  `json:"icon"`
	Autocomplete string    `json:"autocomplete"`   // Recommended
	UID          string    `json:"uid,omitempty"`  // Optional
	Type         string    `json:"type,omitempty"` // Default = "default"
	Mods         Modifiers `json:"mods,omitemtpy"` // Optional
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

// NewVariable create a new variable instacne
func NewVariable(key, value string) Variable {
	return Variable{
		key:   key,
		value: value,
	}
}

// AddIcon adds icon information to the item
func (i Item) AddIcon(iconPath, iconType string) Item {
	i.Icon = IconInfo{
		Type: iconType,
		Path: iconPath,
	}
	return i
}

// Executable make item executable with arg
func (i Item) Executable(arg string) Item {
	i.SubInfo.Arg = arg
	i.SubInfo.Valid = true

	return i
}

// AddOptionalInfo adds optional information
func (i Item) AddOptionalInfo(uid, itemType string) Item {
	i.UID = uid
	i.Type = itemType
	return i
}

// AddVariables adds variables
func (i Item) AddVariables(vars ...Variable) Item {
	return i.addVariables("", vars...)
}

func (i Item) addVariables(key modKey, vars ...Variable) Item {
	var si *SubInfo
	if len(key) == 0 {
		si = &i.SubInfo
	} else {
		switch key {
		case modKeyControl:
			si = &i.Mods.CtrlKey
		case modKeyOption:
			si = &i.Mods.OptionKey
		case modKeyCommand:
			si = &i.Mods.CommandKey
		}
	}
	if si.VarMap == nil {
		si.VarMap = make(map[string]string)
	}
	for _, v := range vars {
		si.VarMap[v.key] = v.value
	}

	return i
}

// AddCtrlKeyAction adds information that will be shown when user pressed the control key
func (i Item) AddCtrlKeyAction(subtitle, arg string, executable bool) Item {
	return i.addModifierAction(modKeyControl, subtitle, arg, executable)
}

// AddCtrlKeyVariables adds variables for ctrl key pressed action
func (i Item) AddCtrlKeyVariables(vars ...Variable) Item {
	return i.addVariables(modKeyControl, vars...)
}

// AddOptionKeyAction adds information that will be shown when user pressed the option key
func (i Item) AddOptionKeyAction(subtitle, arg string, executable bool) Item {
	return i.addModifierAction(modKeyOption, subtitle, arg, executable)
}

// AddOptionKeyVariables adds variables for option key pressed action
func (i Item) AddOptionKeyVariables(vars ...Variable) Item {
	return i.addVariables(modKeyOption, vars...)
}

// AddCommandKeyAction adds information that will be shown when user pressed the command key
func (i Item) AddCommandKeyAction(subtitle, arg string, executable bool) Item {
	return i.addModifierAction(modKeyCommand, subtitle, arg, executable)
}

// AddCommandKeyVariables adds variables for command key pressed action
func (i Item) AddCommandKeyVariables(vars ...Variable) Item {
	return i.addVariables(modKeyCommand, vars...)
}

func (i Item) addModifierAction(key modKey, subtitle, arg string, executable bool) Item {
	var si *SubInfo
	switch key {
	case modKeyControl:
		si = &i.Mods.CtrlKey
	case modKeyOption:
		si = &i.Mods.OptionKey
	case modKeyCommand:
		si = &i.Mods.CommandKey
	}
	if si != nil {
		si.Subtitle = subtitle
		si.Arg = arg
		si.Valid = executable
	}

	return i
}
