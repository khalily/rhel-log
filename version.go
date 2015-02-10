package main

import (
	"sort"
	"strconv"
	"strings"
)

type Version struct {
	major  int
	minor  int
	revise int
	patch  int
	rc1    int
	rc2    int
	str    string
}

type lessFunc func(v1, v2 *Version) bool

type VersionSorter struct {
	versions []*Version
	less     []lessFunc
}

func (vs *VersionSorter) Sort(versions []*Version) {
	vs.versions = versions
	sort.Sort(vs)
}

func OrderedBy(less ...lessFunc) *VersionSorter {
	return &VersionSorter{less: less}
}

func Major(v1, v2 *Version) bool {
	return v1.major < v2.major
}

func Minor(v1, v2 *Version) bool {
	return v1.minor < v2.minor
}

func Revise(v1, v2 *Version) bool {
	return v1.revise < v2.revise
}

func Patch(v1, v2 *Version) bool {
	return v1.patch < v2.patch
}

func Rc1(v1, v2 *Version) bool {
	return v1.rc1 < v2.rc1
}

func Rc2(v1, v2 *Version) bool {
	return v1.rc2 < v2.rc2
}

func (vs *VersionSorter) Less(i, j int) bool {
	p, q := vs.versions[i], vs.versions[j]
	var k int
	for k = 0; k < len(vs.less)-1; k++ {
		less := vs.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return vs.less[k](p, q)
}

func (vs *VersionSorter) Swap(i, j int) {
	vs.versions[i], vs.versions[j] = vs.versions[j], vs.versions[i]
}

func (vs *VersionSorter) Len() int {
	return len(vs.versions)
}

func (v *Version) String() string {
	return v.str
}

func NewVersion(version string) *Version {
	var major int
	var minor int
	var revise int
	var patch int
	var rc1 int
	var rc2 int

	ss := strings.Split(version, ".")
	major, _ = strconv.Atoi(ss[0])
	minor, _ = strconv.Atoi(ss[1])
	if strings.Contains(ss[2], "-") {

		rp := strings.Split(ss[2], "-")
		revise, _ = strconv.Atoi(rp[0])
		patch, _ = strconv.Atoi(rp[1])
	} else {
		revise, _ = strconv.Atoi(ss[2])
	}
	if len(ss) > 4 {
		rc1, _ = strconv.Atoi(ss[3])
	}
	if len(ss) > 5 {
		rc2, _ = strconv.Atoi(ss[4])
	}
	return &Version{major, minor, revise, patch, rc1, rc2, version}
}
