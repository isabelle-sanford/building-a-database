package main

// imports

const INTEGER = 2
const VARCHAR = 3
const INTBYTES = 8 // Integer.BYTES in java

type Schema struct {
	fields map[string]FieldInfo
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

// SCHEMA
func makeSchema() Schema {
	return Schema{make(map[string]FieldInfo)}
}

func (s Schema) addField(fldname string, fldtype int, length int) {
	s.fields[fldname] = FieldInfo{fldtype, length}
}

func (s Schema) addIntField(fldname string) {
	s.addField(fldname, INTEGER, 0)
}

func (s Schema) addStringField(fldname string, length int) {
	s.addField(fldname, VARCHAR, length)
}

func (s Schema) add(fldname string, sch Schema) {
	fldtype := sch.fldtype(fldname)
	length := sch.length(fldname)
	s.addField(fldname, fldtype, length)
}

func (s Schema) addAll(sch Schema) { // ! TEST
	for f := range sch.fields {
		s.add(f, sch)
	}
}

func (s Schema) fldtype(fldname string) int {
	return s.fields[fldname].fldtype
}

func (s Schema) length(fldname string) int {
	return s.fields[fldname].length
}

func (s Schema) hasField(fldname string) bool {
	_, ok := s.fields[fldname]
	return ok
}

// LAYOUTS
func makeLayoutFromSchema(sch Schema) Layout {
	offsets := make(map[string]int)
	pos := INTBYTES // hmm (could be 1? idk)
	for f := range sch.fields {
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
