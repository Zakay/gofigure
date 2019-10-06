package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func lookupIdentifierInRoot(multiKeyName *string) (interface{}, error) {
	keyNames := strings.Split(*multiKeyName, ".")

	var currRoot interface{}
	currRoot = globalRoot
	for i, keyName := range keyNames {
		switch currRoot.(type) {
		case map[string]interface{}:
			if _, exists := currRoot.(map[string]interface{})[keyName]; !exists {
				return nil, errors.New("No key with the name \"" + keyName + "\" exists. From query: \"" + *multiKeyName + "\"")
			}

			currRoot = currRoot.(map[string]interface{})[keyName]
			break
		default:
			return nil, errors.New("Key with the name \"" + keyNames[i] + "\" is not a map. From query: \"" + *multiKeyName + "\"")
		}
	}

	return currRoot, nil
}

func (e *Entry) getFileName() string {
	return e.Pos.Filename
}
func (e *Entry) getLine() int {
	return e.Pos.Line
}
func (e *Entry) getColumn() int {
	return e.Pos.Column
}

func (f *Field) getFileName() string {
	return f.Pos.Filename
}
func (f *Field) getLine() int {
	return f.Pos.Line
}
func (f *Field) getColumn() int {
	return f.Pos.Column
}

func (v *Value) getFileName() string {
	return v.Pos.Filename
}
func (v *Value) getLine() int {
	return v.Pos.Line
}
func (v *Value) getColumn() int {
	return v.Pos.Column
}

type GofigureEntry interface {
	getFileName() string
	getLine() int
	getColumn() int
}

func checkConfigError(err error, v GofigureEntry) {
	if err != nil {
		panic(err.Error() + " in " + v.getFileName() + ":" + strconv.FormatInt(int64(v.getLine()), 10) + ":" + strconv.FormatInt(int64(v.getColumn()), 10))
	}
}

var currPrefix string

func reverseIdentifiers(v *interface{}) {
	switch (*v).(type) {
	case map[string]interface{}:
		for key, value := range (*v).(map[string]interface{}) {
			currPrefixAddition := "." + key

			currPrefix += currPrefixAddition

			reverseIdentifiers(&value)

			currPrefix = string(currPrefix[:len(currPrefix)-len(currPrefixAddition)])
		}

		break
	case []interface{}:
		for _, element := range (*v).([]interface{}) {
			reverseIdentifiers(&element)
		}

		break
	case identifier:
		val, err := lookupIdentifierInRoot((*v).(identifier).name)

		check(err)

		*v = val
		break
	}
}

type identifier struct {
	name *string
}

func (v *Value) toFinalValue() (ret interface{}) {
	if v.Identifier != nil {
		ret = &identifier{v.Identifier}
	} else if v.Map != nil {
		nwMap := map[string]interface{}{}

		for _, field := range v.Map {
			if field.Value == nil {
				nwMap[field.Key] = nil
			} else {
				nwMap[field.Key] = field.Value.toFinalValue()
			}
		}

		ret = nwMap
	} else if v.Float != nil {
		ret = v.Float
	} else if v.Integer != nil {
		ret = v.Integer
	} else if v.Boolean != nil {
		ret = v.Boolean
	} else if v.String != nil {
		ret = v.String
	} else if v.MultilineString != nil {
		ret = v.MultilineString.transform()
	} else if v.FinalArray != nil {
		nwArray := make([]interface{}, len(v.FinalArray), len(v.FinalArray))

		for i, value := range v.FinalArray {
			if value != nil {
				nwArray[i] = value.toFinalValue()
			}
		}

		ret = nwArray
	} else if v.ParsedArray != nil {
		nwArray := make([]interface{}, len(v.ParsedArray), len(v.ParsedArray))

		for i, value := range v.ParsedArray {
			if value != nil {
				nwArray[i] = value.toFinalValue()
			}
		}

		ret = nwArray
	} else { // Has to be empty map
		// todo: include flag for omitting empty values?
		ret = map[string]interface{}{}
	}

	return
}

func mergeMapsOfInterface(dst, src map[string]interface{}) {
	for key, val := range src {
		// Child level select all roots
		if key == "@" {
			for _, root := range dst {
				switch root.(type) {
				case map[string]interface{}:
					mergeMapsOfInterface(root.(map[string]interface{}), val.(map[string]interface{}))
					break
				default:
					// Nothing?
				}
			}

			continue
		}

		if _, exists := dst[key]; !exists {
			dst[key] = val
			continue
		}

		switch src[key].(type) {
		case map[string]interface{}:
			switch dst[key].(type) {
			case map[string]interface{}:
				mergeMapsOfInterface(dst[key].(map[string]interface{}), src[key].(map[string]interface{}))

				break

			default:
				dst[key] = val
			}

			break

		default:
			dst[key] = val
		}
	}
}

func findIdentifierInMap(identifier *string, fields []*Field) (*Value, error) {
	for _, field := range fields {
		value := field.Value

		if field.Key == *identifier {
			return value, nil
		}
	}

	return nil, errors.New("No key called " + *identifier + " exists.")
}

func findIdentifierInConfig(identifier *string, root *FigureConfig) (*Value, error) {
	identParts := strings.Split(*identifier, ".")
	var err error

	for _, entry := range root.Entries {
		field := entry.Field
		value := field.Value

		if len(identParts) == 1 {
			if field.Key == identParts[0] {
				return value, nil
			}

			continue
		} else if value == nil {
			continue
		}

		if field.Key == identParts[0] && value.Map != nil {
			var v *Value
			err = nil

			for j := 1; j < len(identParts) && err == nil; j++ {
				v, err = findIdentifierInMap(&identParts[j], value.Map)

				value = v
			}

			if err == nil {
				return value, err
			}
		}
	}

	return nil, err
}

func reverseIdentifiersInList(values []*Value, root *FigureConfig) {
	for i, value := range values {
		if value.Map != nil {
			reverseIdentifiersInMap(value.Map, root)
		} else if value.ParsedArray != nil {
			reverseIdentifiersInList(value.ParsedArray, root)
		} else if value.Identifier != nil {
			identVal, err := findIdentifierInConfig(value.Identifier, root)
			checkConfigError(err, value)

			values[i] = identVal
		}
	}
}

func reverseIdentifiersInMap(fields []*Field, root *FigureConfig) {
	for _, field := range fields {
		val := field.Value

		if val == nil {
			continue
		}

		if val.Map != nil {
			reverseIdentifiersInMap(val.Map, root)
		} else if val.ParsedArray != nil {
			reverseIdentifiersInList(val.ParsedArray, root)
		} else if val.Identifier != nil {
			identVal, err := findIdentifierInConfig(val.Identifier, root)
			checkConfigError(err, field)

			field.Value = identVal
		}
	}
}

func (c FigureConfig) reverseIdentifiers() {
	for i, entry := range c.Entries {
		field := entry.Field

		if field.Value == nil {
			continue
		}

		// Only look at entries up until this point when searching for identifiers
		tmpConfig := &FigureConfig{Entries: c.Entries[:i+1]}

		if field.Value.Map != nil {
			reverseIdentifiersInMap(field.Value.Map, tmpConfig)
		} else if field.Value.ParsedArray != nil {
			reverseIdentifiersInList(field.Value.ParsedArray, tmpConfig)
		} else if field.Value.Identifier != nil {
			identVal, err := findIdentifierInConfig(field.Value.Identifier, tmpConfig)
			check(err)

			field.Value = identVal
		}
	}
}

func mergeValues(dom, sub *Value) *Value {
	newValue := &Value{Pos: dom.Pos}
	if dom.Map != nil && sub.Map != nil {
		newMap := []*Field{}
		for _, subDominant := range sub.Map {
			var foundField *Field
			for _, dominant := range dom.Map {
				if subDominant.Key == dominant.Key {
					foundField = dominant
					break
				}
			}

			if foundField == nil {
				newMap = append(newMap, subDominant)
			} else {
				newMap = append(newMap, foundField)
			}
		}

		for _, dominant := range dom.Map {
			var foundField bool
			for _, fields := range newMap {
				if fields.Key == dominant.Key {
					foundField = true
					break
				}
			}

			if !foundField {
				newMap = append(newMap, dominant)
			}
		}

		newValue.Map = newMap
	} else {
		newValue = dom
	}

	return newValue
}

func (v Value) mergeArraysWithConfig(prefix string, config *FigureConfig) *Value {
	newValue := &Value{Pos: v.Pos}
	if v.Map != nil {
		newMap := make([]*Field, len(v.Map))
		for i, mapVal := range v.Map {
			newMap[i] = mapVal.mergeArraysWithConfig(prefix, config)
		}
		newValue.Map = newMap
	}

	return newValue
}

func (f Field) mergeArraysWithConfig(prefix string, config *FigureConfig) *Field {
	if f.ArrayIndex != nil {
		val, err := findIdentifierInConfig(&prefix, config)
		check(err)

		var foundField *Field

		for _, mapVal := range val.Map {
			if mapVal.Key == strconv.Itoa(int(*f.ArrayIndex)) {
				foundField = mapVal
				break
			}
		}

		if foundField == nil {
			panic("lol")
		}

		foundField.Value = mergeValues(f.Value, foundField.Value)
		return nil
	} else if f.Value != nil {
		var newPrefix string
		if prefix == "" {
			newPrefix = f.Key
		} else {
			newPrefix = prefix + "." + f.Key
		}

		newValue := &Value{Pos: f.Value.Pos}
		if f.Value.Map != nil {
			newValue.Map = make([]*Field, len(f.Value.Map))
			for i, mapVal := range f.Value.Map {
				newMapVal := mapVal.mergeArraysWithConfig(newPrefix, config)
				if newMapVal == nil {
					return nil
				}
				newValue.Map[i] = newMapVal
			}
		} else {
			newValue = f.Value
		}

		return &Field{Pos: f.Pos, Value: newValue, Key: f.Key}
	}

	return &f
}

func (e Entry) mergeArraysWithConfig(config *FigureConfig) *Entry {
	newField := e.Field.mergeArraysWithConfig("", config)
	if newField == nil {
		return nil
	}

	return &Entry{Pos: e.Pos, Field: newField}
}

func (c FigureConfig) mergeArrays() (ret FigureConfig) {
	ret = FigureConfig{Entries: make([]*Entry, 0)}

	for _, entry := range c.Entries {
		newEntry := entry.mergeArraysWithConfig(&ret)
		if newEntry == nil {
			continue
		}
		ret.Entries = append(ret.Entries, newEntry)
	}

	return
}

func (f *Field) sequentialFieldsToArrays() *Field {
	newField := &Field{
		Pos: f.Pos,
	}

	if f.Value != nil && f.Value.Map != nil {
		newField.Value = &Value{Pos: f.Value.Pos}
		newField.Value.Map = make([]*Field, len(f.Value.Map), len(f.Value.Map))

		for i, mapVal := range f.Value.Map {
			newField.Value.Map[i] = mapVal.sequentialFieldsToArrays()
		}

		if keysAreSequential(newField.Value.Map) {
			arraySize := getLargestKey(newField.Value.Map) + 1
			newField.Value.FinalArray = make([]*Value, arraySize, arraySize)

			for _, mapVal := range newField.Value.Map {
				index, _ := strconv.Atoi(mapVal.Key)
				newField.Value.FinalArray[index] = mapVal.Value
			}

			newField.Value.Map = nil
		}
	} else {
		newField.Value = f.Value
	}

	newField.Key = f.Key

	return newField
}

func (e *Entry) sequentialFieldsToArrays() *Entry {
	newField := e.Field.sequentialFieldsToArrays()

	return &Entry{Pos: e.Pos, Field: newField}
}

func (c FigureConfig) sequentialFieldsToArrays() (ret FigureConfig) {
	ret = FigureConfig{
		Entries: make([]*Entry, len(c.Entries), len(c.Entries)),
		Pos:     c.Pos,
	}

	for i, entry := range c.Entries {
		newEntry := entry.sequentialFieldsToArrays()

		ret.Entries[i] = newEntry
	}

	return
}

// Transform - Takes a parsed and lexed config file and transforms it to a map
func (c FigureConfig) Transform() map[string]interface{} {
	c = c.parseIncludesAndAppendToConfig()
	c = c.explodeSectionsToFields()
	c = c.childFieldsToMap()

	c.reverseIdentifiers()

	c = c.fieldsToArrays()
	c = c.mergeArrays()
	c = c.sequentialFieldsToArrays()

	mapped := c.toMap()

	return mapped
}

func keysAreSequential(arg []*Field) bool {
	var lastIndex int
	for _, mapVal := range arg {
		if arrayIndex, err := strconv.Atoi(mapVal.Key); err == nil {
			if lastIndex < arrayIndex || lastIndex == 0 && arrayIndex == 0 {
				lastIndex = arrayIndex
			} else {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func getLargestKey(arg []*Field) int {
	var largest int
	for _, mapVal := range arg {
		arrayIndex, _ := strconv.Atoi(mapVal.Key)
		if arrayIndex > largest {
			largest = arrayIndex
		}
	}

	return largest
}

func (value *Value) fieldsToArrays() (ret *Value) {
	ret = &Value{Pos: value.Pos}

	if value.Map != nil {
		newMap := make([]*Field, len(value.Map))
		for i, mapVal := range value.Map {
			newMap[i] = mapVal.fieldsToArrays()
		}
		ret.Map = newMap

	} else {
		return value
	}

	return
}

func (field *Field) fieldsToArrays() (ret *Field) {
	if field.Value == nil {
		return field
	}

	ret = &Field{Pos: field.Pos}
	ret.Key = field.Key
	ret.ArrayIndex = field.ArrayIndex
	ret.Value = field.Value.fieldsToArrays()

	return
}

func (config FigureConfig) fieldsToArrays() (ret FigureConfig) {
	ret = FigureConfig{}
	ret.Entries = make([]*Entry, len(config.Entries))

	for i, entry := range config.Entries {
		newField := entry.Field.fieldsToArrays()

		newEntry := &Entry{}
		newEntry.Pos = entry.Pos
		newEntry.Field = newField

		ret.Entries[i] = newEntry
	}

	return
}

var globalRoot map[string]interface{}

func (c FigureConfig) toMap() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	globalRoot = ret

	for _, entry := range c.Entries {
		field := entry.Field
		value := field.Value

		var finalValue interface{}

		if value != nil {
			finalValue = value.toFinalValue()
		}

		// Top level selection of all roots
		if field.Key == "@" {
			var processRoots []map[string]interface{}

			for _, val := range ret {
				switch val.(type) {
				case map[string]interface{}:
					processRoots = append(processRoots, val.(map[string]interface{}))
					break

				default:
					// do nothing?
				}
			}

			for _, root := range processRoots {
				mergeMapsOfInterface(root, finalValue.(map[string]interface{}))
			}

			continue
		}

		// Otherwise regular value
		if value != nil {
			if _, exists := ret[field.Key]; !exists {
				ret[field.Key] = finalValue
			} else {
				switch finalValue.(type) {
				case map[string]interface{}:
					mergeMapsOfInterface(ret[field.Key].(map[string]interface{}), finalValue.(map[string]interface{}))
					break
				default: // finalValue is an array?
					currField := ret[field.Key]
					for _, arrayVal := range finalValue.([]interface{}) {
						currField = append(currField.([]Value), arrayVal.(Value))
					}
					break
				}
			}
		} else {
			ret[field.Key] = nil
		}
	}

	return
}

func (c FigureConfig) childFieldsToMap() (ret FigureConfig) {
	ret = FigureConfig{}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, entry := range c.Entries {
		if entry.Field == nil {
			ret.Entries[i] = entry
			continue
		}

		currField := entry.Field
		for currField.Child != nil {
			currField.Value = &Value{
				Map: []*Field{
					&Field{
						ArrayIndex: currField.Child.ArrayIndex,
						Child:      currField.Child.Child,
						Key:        currField.Child.Key,
						Value:      currField.Child.Value,
						Pos:        currField.Child.Pos,
					}}}
			currField.Child = nil
			currField = currField.Value.Map[0]
		}

		ret.Entries[i] = entry
	}

	return
}

func (s *SectionChild) expandToFields(setTo []*Field) (retVal []*Field) {
	var childFields []*Field

	if s.Child != nil {
		childFields = s.Child.expandToFields(setTo)
	} else {
		childFields = setTo
	}

	// Check for integer range expand macro
	if len(s.Identifier) == 3 && s.Identifier[1] == "..." {
		int1, _ := strconv.Atoi(s.Identifier[0])
		int2, _ := strconv.Atoi(s.Identifier[2])

		length := (int2 - int1) + 1
		newIdentifiers := make([]string, length, length)

		for i, _ := range newIdentifiers {
			newIdentifiers[i] = strconv.Itoa(int1 + i)
		}

		s.Identifier = newIdentifiers
	}

	for _, sectName := range s.Identifier {
		newField := &Field{Key: sectName, Value: &Value{Map: childFields}}

		retVal = append(retVal, newField)
	}

	return
}

func (s *SectionRoot) expandToFields(setTo []*Field) (retVal []*Field) {
	var childFields []*Field

	hasChildren := s.Child != nil

	if hasChildren {
		childFields = s.Child.expandToFields(setTo)
	}

	for _, sectName := range s.Identifier {
		newField := &Field{Key: sectName, Value: &Value{}}

		if !hasChildren {
			newField.Value.Map = setTo
		} else {
			newField.Value.Map = childFields
		}

		retVal = append(retVal, newField)
	}

	return
}

func (s *Section) expandToFields() (retVal []*Field) {
	for _, sectRoot := range s.Roots {
		retVal = append(retVal, sectRoot.expandToFields(s.Fields)...)
	}

	return
}

func (c FigureConfig) explodeSectionsToFields() (ret FigureConfig) {
	ret = FigureConfig{}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, newEntriesIndex := 0, 0; i < len(c.Entries); i++ {
		entry := c.Entries[i]
		if entry.Section == nil {
			ret.Entries[newEntriesIndex] = entry
			newEntriesIndex++
			continue
		}

		section := entry.Section

		newFields := section.expandToFields()
		newEntries := make([]*Entry, len(newFields))

		for i, newField := range newFields {
			newEntries[i] = &Entry{Field: newField}
		}

		firstSlice := ret.Entries[:newEntriesIndex]
		var lastSlice []*Entry

		if newEntriesIndex < len(ret.Entries)-1 {
			lastSlice = ret.Entries[newEntriesIndex+1:]
		}

		ret.Entries = append(firstSlice, newEntries...)
		ret.Entries = append(ret.Entries, lastSlice...)

		newEntriesIndex += len(newEntries)
	}

	return
}

func (c FigureConfig) parseIncludesAndAppendToConfig() (ret FigureConfig) {
	ret = FigureConfig{Pos: c.Pos}
	ret.Entries = make([]*Entry, len(c.Entries))

	for i, newEntriesIndex := 0, 0; i < len(c.Entries); i++ {
		entry := c.Entries[i]

		if entry.Include == nil {
			ret.Entries[newEntriesIndex] = entry
			newEntriesIndex++
			continue
		}

		include := entry.Include
		newConfigList := make([]FigureConfig, len(include.Includes))

		parser := BuildParser()

		// Parse includes
		for j, includeName := range include.Includes {
			newConfig := ParseFile(includeName, parser)

			newConfig = newConfig.parseIncludesAndAppendToConfig()

			newConfigList[j] = newConfig
		}

		// Append new entries in main config and remove the include entry
		for _, newConfig := range newConfigList {
			newEntryList := append(ret.Entries[:newEntriesIndex], newConfig.Entries...)
			newEntryList = append(newEntryList, ret.Entries[newEntriesIndex+1:]...)

			ret.Entries = newEntryList
			newEntriesIndex += len(newConfig.Entries)
		}
	}

	return
}

// This function removes leading/trailing whitespaces, string quotes etc.
func (thisArg *UnprocessedString) transform() (final string) {
	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_leadclose_quotes := regexp.MustCompile(`^("""|''')|("""|''')$`)
	re_inside_whtsp := regexp.MustCompile(`[\r\f\t \p{Zs}]{2,}`)
	re_backslashes := regexp.MustCompile(`\\(?P<C>[^n])`)
	re_newline_whtsp := regexp.MustCompile(`\n +|\n`)

	final = re_leadclose_whtsp.ReplaceAllString(*thisArg.String, "")
	final = re_leadclose_quotes.ReplaceAllString(final, "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")
	final = re_backslashes.ReplaceAllString(final, `\\$C`)
	final = re_newline_whtsp.ReplaceAllString(final, `\\n`)

	return
}
