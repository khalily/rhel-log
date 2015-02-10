package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	prefix      = "https://oss.oracle.com"
	identifer1  = "The rest of"
	identifer2  = "RHEL kernel"
	nbspace     = string([]byte{194, 160})
	rhelKernels = make(map[string]*RhelKernel)
)

type Branch struct {
	Name   string
	LogUri string
}

type BranchList []*Branch

type RhelKernel struct {
	Version     string
	Description string
	SubSystems  map[string]*SubSystem
}

type SubSystem struct {
	Name    string
	Drivers []*Driver
}

type Driver struct {
	Name      string
	ChangeLog string
}

type DocumentManager struct {
	RhelKernels map[string]*RhelKernel
	cur_rk      *RhelKernel
}

func (dm *DocumentManager) fetchOneBranch(b *Branch) {
	doc, err := goquery.NewDocument(prefix + b.LogUri)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".log_body").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		text = strings.Replace(text, string(nbspace), " ", -1)
		text = strings.TrimLeft(text, "\n")

		if strings.Contains(text, identifer1) {
			dm.ParseOnePage(text)
		} else if strings.Contains(text, identifer2) {
			dm.ParseOnePage(text)
		}
	})

}

func (rk *RhelKernel) String() (str string) {
	str += fmt.Sprintf("Version: %s\nDescription: %s\nSubsystem:\n", rk.Version, rk.Description)
	for _, v := range rk.SubSystems {
		str += fmt.Sprintf("\t%s:\n", v.Name)
		for _, dri := range v.Drivers {
			str += fmt.Sprintf("\t\t[%s] %s\n", dri.Name, dri.ChangeLog)
		}
	}
	return
}

func (dm *DocumentManager) String() (str string) {
	for _, v := range dm.RhelKernels {
		str += v.String() + "\n"
	}
	return
}

func (dm *DocumentManager) parseOneLine(line string) {
	re_kernel, err := regexp.Compile("^\\* (.*) \\[(.*)\\]")
	re_subsystem, err := regexp.Compile("^- \\[(.*)\\] (.*): (.*)")
	if err != nil {
		log.Fatal(err)
	}

	rk := dm.cur_rk
	if re_kernel.MatchString(line) {
		if rk != nil {
			if _, ok := dm.RhelKernels[rk.Version]; !ok {
				dm.RhelKernels[rk.Version] = rk
			}
		}
		dm.cur_rk = &RhelKernel{SubSystems: make(map[string]*SubSystem)}
		rk = dm.cur_rk

		substrs := re_kernel.FindAllStringSubmatch(line, -1)

		rk.Description = substrs[0][1]
		rk.Version = substrs[0][2]

	} else if re_subsystem.MatchString(line) {
		substrs := re_subsystem.FindAllStringSubmatch(line, -1)

		subsys_name := substrs[0][1]
		if _, ok := rk.SubSystems[subsys_name]; !ok {
			rk.SubSystems[subsys_name] = &SubSystem{Name: subsys_name}
		}

		driver_name := substrs[0][2]
		change_log := substrs[0][3]
		driver := &Driver{Name: driver_name, ChangeLog: change_log}

		rk.SubSystems[subsys_name].Drivers =
			append(rk.SubSystems[subsys_name].Drivers, driver)
	}
}

func (dm *DocumentManager) ParseOnePage(text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) > 0 {
			dm.parseOneLine(line)
		}
	}
}

func main() {
	doc, err := goquery.NewDocument(prefix + "/git/?p=redpatch.git;a=heads")
	if err != nil {
		log.Fatal(err)
	}

	var bl BranchList

	doc.Find(".heads tr").Each(func(i int, s *goquery.Selection) {
		name := s.Find(".name").Text()
		uri, ext := s.Find(".link a").Eq(1).Attr("href")
		if !ext {
			log.Printf("not found uri of name %s\n", name)
			uri = ""
		}
		branch := &Branch{Name: name, LogUri: uri}
		bl = append(bl, branch)
	})

	dm := &DocumentManager{RhelKernels: make(map[string]*RhelKernel)}

	for _, v := range bl {
		dm.fetchOneBranch(v)
	}

	var versions []*Version
	for version, _ := range dm.RhelKernels {
		versions = append(versions, NewVersion(version))
	}
	vs := OrderedBy(Major, Minor, Revise, Patch, Rc1, Rc2)
	vs.Sort(versions)

	for _, v := range versions {
		fmt.Println(dm.RhelKernels[v.String()])
	}
}
