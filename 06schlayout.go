package main

import "fmt"

// imports

const INTEGER = 2
const VARCHAR = 3
const INTBYTES = 8 // Integer.BYTES in java

type Schema struct {
	fields    map[string]FieldInfo
	fieldlist []string
}

type FieldInfo struct {
	fldtype int
	length  int
}

type Layout struct {
	schema   Schema
	offsets  map[string]int
	slotsize int
}

func (fi FieldInfo) String() string {
	if fi.fldtype == VARCHAR {
		return fmt.Sprint("VARCHAR(", fi.length, ")")
	} else { // field is integer
		return "INTEGER"
	}
}

func (s Schema) String() string {
	ret := "Schema: "
	for k, v := range s.fields {
		ret += k + " [" + v.String() + "] "
	}
	return ret
}

func (l Layout) String() string {
	ret := fmt.Sprint("Layout (slot size ", l.slotsize, "): ")

	for fldname, offset := range l.offsets {
		fldinfo := l.schema.fields[fldname]
		ret += fmt.Sprint(fldname, "{", fldinfo, "[", offset, "]} ")
	}

	return ret
}

// SCHEMA
// could just do var sch Schema here
func makeSchema() Schema {
	return Schema{make(map[string]FieldInfo), make([]string, 0)}
}

func (s *Schema) addField(fldname string, fldtype int, length int) {
	s.fields[fldname] = FieldInfo{fldtype, length}
	s.fieldlist = append(s.fieldlist, fldname)
	//fmt.Println("adding field ", fldname, "fldlist is now ", s.fieldlist)
}

func (s *Schema) addIntField(fldname string) {
	s.addField(fldname, INTEGER, 0)
}

func (s *Schema) addStringField(fldname string, length int) {
	s.addField(fldname, VARCHAR, length)
}

func (s *Schema) add(fldname string, sch Schema) {
	fldtype := sch.fldtype(fldname)
	length := sch.length(fldname)
	s.addField(fldname, fldtype, length)
}

func (s *Schema) addAll(sch Schema) { // ! TEST
	for f := range sch.fields {
		s.add(f, sch)
		s.fieldlist = sch.fieldlist
	}
}

func (s *Schema) fldtype(fldname string) int {
	return s.fields[fldname].fldtype
}

func (s *Schema) length(fldname string) int {
	return s.fields[fldname].length
}

func (s *Schema) hasField(fldname string) bool {
	_, ok := s.fields[fldname]
	return ok
}

// LAYOUTS
func makeLayoutFromSchema(sch Schema) Layout {
	offsets := make(map[string]int)
	pos := INTBYTES // hmm (could be 1? idk)
	for _, f := range sch.fieldlist {
		offsets[f] = pos
		pos += sch.lengthInBytes(f)
	}
	slotsize := pos
	var l = Layout{sch, offsets, slotsize}

	return l
}

// is this necessary? standard constructor is just as easy
func makeLayout(sch Schema, offsets map[string]int, slotsize int) Layout {
	var l = Layout{sch, offsets, slotsize}
	return l
}

func (sch Schema) lengthInBytes(fldname string) int {
	fldtype := sch.fldtype(fldname)
	if fldtype == INTEGER {
		return INTBYTES
	} else { // i.e. fldtype == VARCHAR
		return sch.length(fldname) + INTBYTES // hmmmmm
	}
}
