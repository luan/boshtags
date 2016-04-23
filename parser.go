package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// tagParser contains the data needed while parsing.
type tagParser struct {
	file     string
	tags     []Tag    // list of created tags
	types    []string // all types we encounter, used to determine the constructors
	relative bool     // should filenames be relative to basepath
	basepath string   // output file directory
}

// Parse parses the source in filename and returns a list of tags. If relative
// is true, the filenames in the list of tags are relative to basepath.
func Parse(filename string, relative bool, basepath string) ([]Tag, error) {
	p := &tagParser{
		tags:     []Tag{},
		types:    make([]string, 0),
		relative: relative,
		basepath: basepath,
		file:     filename,
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// declarations
	p.parse(f)

	return p.tags, nil
}

// parseDeclarations creates a tag for each function, type or value declaration.
func (p *tagParser) parse(f io.Reader) {
	r := bufio.NewReader(f)
	var currentSection TagType
	var currentPrimitive, currentInstanceGroup, prefix string
	jobsPrefix := "####"
	lineNum := 0
	var nextLine *string
	inStemcells := false

	nl, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read input file: %s\n", err)
		return
	}
	nextLine = &nl

	for {
		lineNum++
		if nextLine == nil {
			break
		}
		line := *nextLine
		nl, err := r.ReadString('\n')
		if err == io.EOF {
			nextLine = nil
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't read input file: %s\n", err)
			return
		}
		nextLine = &nl
		sectionRegexp := regexp.MustCompile("^(name|director_uuid|vm_types|resource_pools|disk_types|disk_pools|networks|azs|update|compilation|instance_groups|releases|stemcells): ?(.*)")
		sectionMatch := sectionRegexp.FindStringSubmatch(line)
		if len(sectionMatch) > 0 {
			name := sectionMatch[1]
			inStemcells = false
			switch name {
			case "vm_types":
				currentSection = VM
			case "resource_pools":
				currentSection = VM
				name = "vm_types"
			case "disk_types":
				currentSection = Disk
			case "disk_pools":
				currentSection = Disk
				name = "disk_types"
			case "networks":
				currentSection = Network
			case "azs":
				currentSection = AZ
			case "instance_groups":
				currentSection = InstanceGroup
			case "releases":
				currentSection = Release
			case "stemcells":
				currentSection = Stemcell
				inStemcells = true
			}
			if nextLine != nil {
				index := strings.IndexAny(*nextLine, "-abcdefghijklmnopqrstuvwxyz")
				if index > 0 {
					prefix = (*nextLine)[0:index]
				} else {
					prefix = ""
				}
			}
			currentPrimitive = name

			addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
			tag := p.createTag(name, addr, lineNum, Primitive)
			tag.Fields[TypeField] = "section"
			if name == "name" {
				tag.Type = Basic
				tag.Fields[TypeField] = sectionMatch[2]
			}
			if name == "director_uuid" {
				tag.Type = Basic
				tag.Fields[TypeField] = sectionMatch[2]
			}
			p.tags = append(p.tags, tag)
			continue
		}

		if !inStemcells {
			lineRegexp := regexp.MustCompile("^" + prefix + "[- ] name: (.*)")
			lineMatch := lineRegexp.FindStringSubmatch(line)
			if len(lineMatch) > 0 {
				name := lineMatch[1]
				addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
				tag := p.createTag(name, addr, lineNum, currentSection)
				tag.Fields[PrimitiveType] = currentPrimitive
				tag.Fields[TypeField] = currentPrimitive
				currentInstanceGroup = name
				p.tags = append(p.tags, tag)
				continue
			}
		} else {
			lineRegexp := regexp.MustCompile("^" + prefix + "[- ] alias: (.*)")
			lineMatch := lineRegexp.FindStringSubmatch(line)
			if len(lineMatch) > 0 {
				name := lineMatch[1]
				addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
				tag := p.createTag(name, addr, lineNum, currentSection)
				tag.Fields[PrimitiveType] = currentPrimitive
				tag.Fields[TypeField] = currentPrimitive
				currentInstanceGroup = name
				p.tags = append(p.tags, tag)
				continue
			}
		}

		lineRegexp := regexp.MustCompile("^" + jobsPrefix + "[- ] name: (.*)")
		lineMatch := lineRegexp.FindStringSubmatch(line)
		if len(lineMatch) > 0 {
			name := lineMatch[1]
			addr := fmt.Sprintf("/\\%%%dl\\%%%dc/", lineNum, strings.Index(string(line), name)+1)
			tag := p.createTag(name, addr, lineNum, Job)
			tag.Fields[InstanceGroupType] = currentPrimitive + "." + currentInstanceGroup
			tag.Fields[TypeField] = "job"
			p.tags = append(p.tags, tag)
			continue
		}

		lineRegexp = regexp.MustCompile("^" + prefix + "[- ] jobs:")
		if lineRegexp.MatchString(line) {
			if nextLine != nil {
				index := strings.IndexAny(*nextLine, "-abcdefghijklmnopqrstuvwxyz")
				jobsPrefix = (*nextLine)[0:index]
			}
			continue
		}
	}
}

// createTag creates a new tag, using pos to find the filename and set the line number.
func (p *tagParser) createTag(name, addr string, line int, tagType TagType) Tag {
	f := p.file
	if p.relative {
		if abs, err := filepath.Abs(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not determine absolute path: %s\n", err)
		} else if rel, err := filepath.Rel(p.basepath, abs); err != nil {
			fmt.Fprintf(os.Stderr, "could not determine relative path: %s\n", err)
		} else {
			f = rel
		}
	}
	return NewTag(name, f, addr, line, tagType)
}
