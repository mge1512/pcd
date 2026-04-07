// Package lint implements the pcd-lint specification validation logic.
// SPDX-License-Identifier: GPL-2.0-only

package lint

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ── Version constants ─────────────────────────────────────────────────────────

const (
	SpecSchema = "0.3.21"
)

// ── Types ─────────────────────────────────────────────────────────────────────

// Severity represents the severity of a diagnostic.
type Severity int

const (
	SevError   Severity = iota
	SevWarning
)

func (s Severity) String() string {
	if s == SevError {
		return "error"
	}
	return "warning"
}

// Diagnostic represents a single lint finding.
type Diagnostic struct {
	Severity Severity
	Section  string
	Message  string
	Rule     string
	Line     int
}

// LintResult is the result of linting a spec.
type LintResult struct {
	File        string
	Valid        bool
	Errors       int
	Warnings     int
	Diagnostics []Diagnostic
}

// ── SPDX license list ─────────────────────────────────────────────────────────

var spdxLicenses = map[string]bool{
	"0BSD": true, "AAL": true, "Abstyles": true, "AdaCore-doc": true,
	"Adobe-2006": true, "Adobe-Glyph": true, "ADSL": true, "AFL-1.1": true,
	"AFL-1.2": true, "AFL-2.0": true, "AFL-2.1": true, "AFL-3.0": true,
	"Agentejo": true, "AGPL-1.0-only": true, "AGPL-1.0-or-later": true,
	"AGPL-3.0-only": true, "AGPL-3.0-or-later": true, "Aladdin": true,
	"AMDPLPA": true, "AML": true, "AMPAS": true, "ANTLR-PD": true,
	"ANTLR-PD-fallback": true, "Apache-1.0": true, "Apache-1.1": true,
	"Apache-2.0": true, "APAFML": true, "APL-1.0": true, "App-s2p": true,
	"APSL-1.0": true, "APSL-1.1": true, "APSL-1.2": true, "APSL-2.0": true,
	"Arphic-1999": true, "Artistic-1.0": true, "Artistic-1.0-cl8": true,
	"Artistic-1.0-Perl": true, "Artistic-2.0": true,
	"ASWF-Digital-Assets-1.0": true, "ASWF-Digital-Assets-1.1": true,
	"Beerware": true, "Bitstream-Charter": true, "Bitstream-Vera": true,
	"BitTorrent-1.0": true, "BitTorrent-1.1": true, "blessing": true,
	"BlueOak-1.0.0": true, "Borceux": true, "BSD-1-Clause": true,
	"BSD-2-Clause": true, "BSD-2-Clause-Patent": true, "BSD-2-Clause-Views": true,
	"BSD-3-Clause": true, "BSD-3-Clause-Clear": true, "BSD-3-Clause-LBNL": true,
	"BSD-3-Clause-Modification": true, "BSD-3-Clause-No-Nuclear-License": true,
	"BSD-3-Clause-No-Nuclear-License-2014": true,
	"BSD-3-Clause-No-Nuclear-Warranty": true, "BSD-3-Clause-Open-MPI": true,
	"BSD-4-Clause": true, "BSD-4-Clause-Shortened": true, "BSD-4-Clause-UC": true,
	"BSD-4.3RENO": true, "BSD-4.3TAHOE": true,
	"BSD-Advertising-Acknowledgement": true,
	"BSD-Attribution-HPND-disclaimer": true, "BSD-Protection": true,
	"BSD-Source-Code": true, "BSL-1.0": true, "BUSL-1.1": true,
	"CAL-1.0": true, "CAL-1.0-Combined-Work-Exception": true, "Caldera": true,
	"CATOSL-1.1": true, "CC-BY-1.0": true, "CC-BY-2.0": true,
	"CC-BY-2.5": true, "CC-BY-2.5-AU": true, "CC-BY-3.0": true,
	"CC-BY-3.0-AT": true, "CC-BY-3.0-DE": true, "CC-BY-3.0-IGO": true,
	"CC-BY-3.0-NL": true, "CC-BY-3.0-US": true, "CC-BY-4.0": true,
	"CC-BY-NC-1.0": true, "CC-BY-NC-2.0": true, "CC-BY-NC-2.5": true,
	"CC-BY-NC-3.0": true, "CC-BY-NC-4.0": true, "CC-BY-NC-ND-1.0": true,
	"CC-BY-NC-ND-2.0": true, "CC-BY-NC-ND-2.5": true, "CC-BY-NC-ND-3.0": true,
	"CC-BY-NC-ND-3.0-IGO": true, "CC-BY-NC-ND-4.0": true,
	"CC-BY-NC-SA-1.0": true, "CC-BY-NC-SA-2.0": true,
	"CC-BY-NC-SA-2.0-DE": true, "CC-BY-NC-SA-2.0-FR": true,
	"CC-BY-NC-SA-2.0-UK": true, "CC-BY-NC-SA-2.5": true,
	"CC-BY-NC-SA-3.0": true, "CC-BY-NC-SA-3.0-IGO": true,
	"CC-BY-NC-SA-4.0": true, "CC-BY-ND-1.0": true, "CC-BY-ND-2.0": true,
	"CC-BY-ND-2.5": true, "CC-BY-ND-3.0": true, "CC-BY-ND-3.0-IGO": true,
	"CC-BY-ND-4.0": true, "CC-BY-SA-1.0": true, "CC-BY-SA-2.0": true,
	"CC-BY-SA-2.0-UK": true, "CC-BY-SA-2.1-JP": true, "CC-BY-SA-2.5": true,
	"CC-BY-SA-3.0": true, "CC-BY-SA-3.0-AT": true, "CC-BY-SA-3.0-IGO": true,
	"CC-BY-SA-4.0": true, "CC-PDDC": true, "CC0-1.0": true, "CDDL-1.0": true,
	"CDDL-1.1": true, "CDL-1.0": true, "CDLA-Permissive-1.0": true,
	"CDLA-Permissive-2.0": true, "CDLA-Sharing-1.0": true, "CECILL-1.0": true,
	"CECILL-1.1": true, "CECILL-2.0": true, "CECILL-2.1": true,
	"CECILL-B": true, "CECILL-C": true, "CERN-OHL-P-2.0": true,
	"CERN-OHL-S-2.0": true, "CERN-OHL-W-2.0": true, "CFITSIO": true,
	"checkmk": true, "ClArtistic": true, "Clips": true, "CMU-Mach": true,
	"CNRI-Jython": true, "CNRI-Python": true,
	"CNRI-Python-GPL-Compatible": true, "COIL-1.0": true,
	"Community-Spec-1.0": true, "Condor-1.1": true,
	"copyleft-next-0.3.0": true, "copyleft-next-0.3.1": true,
	"Cornell-Lossless-JPEG": true, "CPAL-1.0": true, "CPL-1.0": true,
	"CPOL-1.02": true, "Crossword": true, "CrystalStacker": true,
	"CUA-OPL-1.0": true, "Cube": true, "curl": true, "D-FSL-1.0": true,
	"diffmark": true, "DL-DE-BY-2.0": true, "DOC": true, "Dotseqn": true,
	"DRL-1.0": true, "DSDP": true, "dtoa": true, "dvipdfm": true,
	"ECL-1.0": true, "ECL-2.0": true, "EFL-1.0": true, "EFL-2.0": true,
	"elastic-2.0": true, "Entessa": true, "EPICS": true, "EPL-1.0": true,
	"EPL-2.0": true, "ErlPL-1.1": true, "etalab-2.0": true,
	"EUDatagrid": true, "EUPL-1.0": true, "EUPL-1.1": true, "EUPL-1.2": true,
	"Eurosym": true, "Fair": true, "FDK-AAC": true, "Frameworx-1.0": true,
	"FreeBSD-DOC": true, "FreeImage": true, "FSFAP": true, "FSFUL": true,
	"FSFULLWD": true, "FTL": true, "GD": true,
	"GFDL-1.1-invariants-only": true, "GFDL-1.1-invariants-or-later": true,
	"GFDL-1.1-no-invariants-only": true,
	"GFDL-1.1-no-invariants-or-later": true, "GFDL-1.1-only": true,
	"GFDL-1.1-or-later": true, "GFDL-1.2-invariants-only": true,
	"GFDL-1.2-invariants-or-later": true,
	"GFDL-1.2-no-invariants-only": true,
	"GFDL-1.2-no-invariants-or-later": true, "GFDL-1.2-only": true,
	"GFDL-1.2-or-later": true, "GFDL-1.3-invariants-only": true,
	"GFDL-1.3-invariants-or-later": true,
	"GFDL-1.3-no-invariants-only": true,
	"GFDL-1.3-no-invariants-or-later": true, "GFDL-1.3-only": true,
	"GFDL-1.3-or-later": true, "Giftware": true, "GL2PS": true,
	"Glide": true, "Glulxe": true, "GLWTPL": true, "gnuplot": true,
	"GPL-2.0-only": true, "GPL-2.0-or-later": true, "GPL-3.0-only": true,
	"GPL-3.0-or-later": true, "Graphics-Gems": true, "gSOAP-1.3b": true,
	"HaskellReport": true, "Hippocratic-2.1": true, "HP-1986": true,
	"HPND": true, "HPND-Markus-Kuhn": true, "HPND-sell-variant": true,
	"HPND-sell-variant-MIT-disclaimer": true, "HTMLTIDY": true,
	"IBM-pibs": true, "ICU": true, "IEC-Code-Components-EULA": true,
	"IJG": true, "IJG-short": true, "ImageMagick": true, "iMatix": true,
	"Imlib2": true, "Info-ZIP": true, "Inner-Net-2.0": true, "Intel": true,
	"Intel-ACPI": true, "Interbase-1.0": true, "IPA": true, "IPL-1.0": true,
	"ISC": true, "Jam": true, "JasPer-2.0": true, "JPL-image": true,
	"JPNIC": true, "JSON": true, "Kazlib": true, "knuth-ctan": true,
	"LGPL-2.0-only": true, "LGPL-2.0-or-later": true, "LGPL-2.1-only": true,
	"LGPL-2.1-or-later": true, "LGPL-3.0-only": true,
	"LGPL-3.0-or-later": true, "LGPLLR": true, "Libpng": true,
	"libpng-2.0": true, "libselinux-1.0": true, "libtiff": true,
	"LiLiQ-P-1.1": true, "LiLiQ-R-1.1": true, "LiLiQ-Rplus-1.1": true,
	"Linux-OpenIB": true, "Linux-TLDP": true,
	"Linux-man-pages-1-para": true, "Linux-man-pages-copyleft": true,
	"Linux-man-pages-copyleft-2-para": true,
	"Linux-man-pages-copyleft-var": true, "LPL-1.0": true, "LPL-1.02": true,
	"LPPL-1.0": true, "LPPL-1.1": true, "LPPL-1.2": true, "LPPL-1.3a": true,
	"LPPL-1.3c": true, "MakeIndex": true, "Martin-Ware": true,
	"Minpack": true, "MirOS": true, "MIT": true, "MIT-0": true,
	"MIT-advertising": true, "MIT-CMU": true, "MIT-enna": true,
	"MIT-feh": true, "MIT-Festival": true, "MIT-Modern-Variant": true,
	"MIT-open-group": true, "MIT-Wu": true, "MITNFA": true, "Motosoto": true,
	"MPL-1.0": true, "MPL-1.1": true, "MPL-2.0": true,
	"MPL-2.0-no-copyleft-exception": true, "mplus": true, "MS-LPL": true,
	"MS-PL": true, "MS-RL": true, "MTLL": true, "MulanPSL-1.0": true,
	"MulanPSL-2.0": true, "Multics": true, "Mup": true, "NAIST-2003": true,
	"NASA-1.3": true, "NBPL-1.0": true, "NCGL-UK-2.0": true, "NCSA": true,
	"Net-SNMP": true, "NetCDF": true, "Newsletr": true, "NGPL": true,
	"NICTA-1.0": true, "NIST-PD": true, "NIST-PD-fallback": true,
	"NIST-Software": true, "NLOD-1.0": true, "NLOD-2.0": true, "NLPL": true,
	"Nokia": true, "NOSL": true, "Noweb": true, "NPL-1.0": true,
	"NPL-1.1": true, "NPOSL-3.0": true, "NRL": true, "NTP": true,
	"NTP-0": true, "O-UDA-1.0": true, "OAR": true, "OCCT-PL": true,
	"OCLC-2.0": true, "ODbL-1.0": true, "ODC-By-1.0": true, "OFFIS": true,
	"OFL-1.0": true, "OFL-1.0-no-RFN": true, "OFL-1.0-RFN": true,
	"OFL-1.1": true, "OFL-1.1-no-RFN": true, "OFL-1.1-RFN": true,
	"OGC-1.0": true, "OGDL-Taiwan-1.0": true, "OGL-Canada-2.0": true,
	"OGL-UK-1.0": true, "OGL-UK-2.0": true, "OGL-UK-3.0": true,
	"OGTSL": true, "OLDAP-1.1": true, "OLDAP-1.2": true, "OLDAP-1.3": true,
	"OLDAP-1.4": true, "OLDAP-2.0": true, "OLDAP-2.0.1": true,
	"OLDAP-2.1": true, "OLDAP-2.2": true, "OLDAP-2.2.1": true,
	"OLDAP-2.2.2": true, "OLDAP-2.3": true, "OLDAP-2.4": true,
	"OLDAP-2.5": true, "OLDAP-2.6": true, "OLDAP-2.7": true,
	"OLDAP-2.8": true, "OML": true, "OpenPBS-2.3": true, "OpenSSL": true,
	"OPL-1.0": true, "OPL-UK-3.0": true, "OPUBL-1.0": true,
	"OSET-PL-2.1": true, "OSL-1.0": true, "OSL-1.1": true, "OSL-2.0": true,
	"OSL-2.1": true, "OSL-3.0": true, "Parity-6.0.0": true,
	"Parity-7.0.0": true, "PDDL-1.0": true, "PHP-3.0": true,
	"PHP-3.01": true, "Plexus-Classworlds": true,
	"PolyForm-Noncommercial-1.0.0": true,
	"PolyForm-Small-Business-1.0.0": true, "PostgreSQL": true,
	"PSF-2.0": true, "psfrag": true, "psutils": true, "Python-2.0": true,
	"Python-2.0.1": true, "Qhull": true, "QPL-1.0": true,
	"QPL-1.0-INRIA-2004": true, "RHeCos-v1.1": true, "RPL-1.1": true,
	"RPL-1.5": true, "RPSL-1.0": true, "RSA-MD": true, "RSCPL": true,
	"Ruby": true, "SAX-PD": true, "Saxpath": true, "SCEA": true,
	"SchemeReport": true, "Sendmail": true, "Sendmail-8.23": true,
	"SGI-B-1.0": true, "SGI-B-1.1": true, "SGI-B-2.0": true, "SGP4": true,
	"SHL-0.5": true, "SHL-0.51": true, "SimPL-2.0": true, "SISSL": true,
	"Sleepycat": true, "SMLNJ": true, "SMPPL": true, "SNIA": true,
	"snprintf": true, "Spencer-86": true, "Spencer-94": true,
	"Spencer-99": true, "SPL-1.0": true, "SSH-OpenSSH": true,
	"SSH-short": true, "SSPL-1.0": true, "StandardML-NJ": true,
	"SugarCRM-1.1.3": true, "SunPro": true, "SWL": true, "Symlinks": true,
	"TAPR-OHL-1.0": true, "TCL": true, "TCP-wrappers": true,
	"TermReadKey": true, "TMate": true, "TORQUE-1.1": true, "TOSL": true,
	"TPDL": true, "TPL-1.0": true, "TTWL": true, "TU-Berlin-1.0": true,
	"TU-Berlin-2.0": true, "UCAR": true, "UCL-2.0": true,
	"Unicode-DFS-2015": true, "Unicode-DFS-2016": true, "Unicode-TOU": true,
	"UnixCrypt": true, "Unlicense": true, "UPL-1.0": true, "Vim": true,
	"VOSTROM": true, "VSL-1.0": true, "W3C": true, "W3C-19980720": true,
	"W3C-20150513": true, "w3m": true, "Watcom-1.0": true,
	"Widget-Workshop": true, "Wsuipa": true, "WTFPL": true,
	"wxWindows": true, "X11": true,
	"X11-distribute-modifications-variant": true, "Xdebug-1.03": true,
	"Xerox": true, "Xfig": true, "XFree86-1.1": true, "xlock": true,
	"Xnet": true, "xpp": true, "XSkat": true, "YPL-1.0": true,
	"YPL-1.1": true, "Zed": true, "Zend-2.0": true, "Zimbra-1.3": true,
	"Zimbra-1.4": true, "Zlib": true, "zlib-acknowledgement": true,
	"ZPL-1.1": true, "ZPL-2.0": true, "ZPL-2.1": true,
}

// IsValidSPDX validates a license identifier or compound expression.
func IsValidSPDX(expr string) bool {
	parts := regexp.MustCompile(`\s+(?:OR|AND|WITH)\s+`).Split(expr, -1)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimPrefix(p, "(")
		p = strings.TrimSuffix(p, ")")
		p = strings.TrimSpace(p)
		if !spdxLicenses[p] {
			return false
		}
	}
	return len(parts) > 0
}

// ── Known deployment templates ────────────────────────────────────────────────

var KnownTemplates = []string{
	"wasm", "ebpf", "kernel-module", "verified-library",
	"cli-tool", "gui-tool", "cloud-native", "backend-service",
	"library-c-abi", "enterprise-software", "academic",
	"python-tool", "enhance-existing", "manual", "template",
	"mcp-server", "project-manifest",
}

// ── Parsed spec structures ────────────────────────────────────────────────────

type specLine struct {
	num  int
	text string
}

type parsedSpec struct {
	lines         []specLine
	sections      map[string]int
	metaFields    map[string]string
	metaAuthors   []string
	metaLineNums  map[string]int
	authorLineNum int
	behaviorNames []string
	behaviorLines map[string]int
	milestones    []milestone
}

type milestone struct {
	name            string
	lineNum         int
	status          string
	statusLineNum   int
	scaffold        string
	scaffoldLineNum int
	includedBehavs  []string
	deferredBehavs  []string
	hasIncluded     bool
	hasDeferred     bool
	hasAcceptance   bool
	hasStatus       bool
	hasScaffold     bool
}

var (
	reBehavior        = regexp.MustCompile(`^## BEHAVIOR(?:/INTERNAL)?: (.+)$`)
	reMilestone       = regexp.MustCompile(`^## MILESTONE: (.+)$`)
	reMetaField       = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*): +(.*)$`)
	reSemanticVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	reTypeDecl        = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]*) :=`)
)

func parseSpecFromString(content string) (*parsedSpec, error) {
	ps := &parsedSpec{
		sections:      make(map[string]int),
		metaFields:    make(map[string]string),
		metaLineNums:  make(map[string]int),
		behaviorLines: make(map[string]int),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	fenceDepth := 0
	inMeta := false
	inMilestone := false
	currentMilestone := milestone{}

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		ps.lines = append(ps.lines, specLine{num: lineNum, text: raw})

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		if strings.HasPrefix(raw, "## ") {
			heading := strings.TrimSpace(raw)
			ps.sections[heading] = lineNum

			inMeta = (heading == "## META")

			if inMilestone {
				ps.milestones = append(ps.milestones, currentMilestone)
				currentMilestone = milestone{}
				inMilestone = false
			}

			if m := reBehavior.FindStringSubmatch(raw); m != nil {
				name := strings.TrimSpace(m[1])
				ps.behaviorNames = append(ps.behaviorNames, name)
				ps.behaviorLines[name] = lineNum
			}

			if m := reMilestone.FindStringSubmatch(raw); m != nil {
				inMilestone = true
				currentMilestone = milestone{
					name:    strings.TrimSpace(m[1]),
					lineNum: lineNum,
				}
			}

			continue
		}

		if inMeta {
			if m := reMetaField.FindStringSubmatch(raw); m != nil {
				key := m[1]
				val := strings.TrimSpace(m[2])
				if key == "Author" {
					if ps.authorLineNum == 0 {
						ps.authorLineNum = lineNum
					}
					ps.metaAuthors = append(ps.metaAuthors, val)
				} else {
					if _, exists := ps.metaFields[key]; !exists {
						ps.metaLineNums[key] = lineNum
					}
					ps.metaFields[key] = val
				}
			}
		}

		if inMilestone {
			if strings.HasPrefix(raw, "Status:") {
				currentMilestone.hasStatus = true
				currentMilestone.statusLineNum = lineNum
				parts := strings.SplitN(raw, ":", 2)
				if len(parts) == 2 {
					currentMilestone.status = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(raw, "Scaffold:") {
				currentMilestone.hasScaffold = true
				currentMilestone.scaffoldLineNum = lineNum
				parts := strings.SplitN(raw, ":", 2)
				if len(parts) == 2 {
					currentMilestone.scaffold = strings.TrimSpace(parts[1])
				}
			} else if strings.HasPrefix(raw, "Included BEHAVIORs:") {
				currentMilestone.hasIncluded = true
				parts := strings.SplitN(raw, ":", 2)
				if len(parts) == 2 {
					currentMilestone.includedBehavs = splitBehaviorList(parts[1])
				}
			} else if strings.HasPrefix(raw, "Deferred BEHAVIORs:") {
				currentMilestone.hasDeferred = true
				parts := strings.SplitN(raw, ":", 2)
				if len(parts) == 2 {
					currentMilestone.deferredBehavs = splitBehaviorList(parts[1])
				}
			} else if strings.HasPrefix(raw, "Acceptance criteria:") {
				currentMilestone.hasAcceptance = true
			}
		}
	}

	if inMilestone {
		ps.milestones = append(ps.milestones, currentMilestone)
	}

	return ps, scanner.Err()
}

func splitBehaviorList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func (ps *parsedSpec) hasBehaviorSection() bool {
	for k := range ps.sections {
		if strings.HasPrefix(k, "## BEHAVIOR: ") || strings.HasPrefix(k, "## BEHAVIOR/INTERNAL: ") {
			return true
		}
	}
	return false
}

func (ps *parsedSpec) linesInSection(sectionHeading string) []specLine {
	startLine, ok := ps.sections[sectionHeading]
	if !ok {
		return nil
	}

	var result []specLine
	inSection := false
	fd := 0

	for _, sl := range ps.lines {
		if sl.num == startLine {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}

		trimmed := strings.TrimSpace(sl.text)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fd == 0 {
				fd = 1
			} else {
				fd--
			}
			continue
		}
		if fd > 0 {
			continue
		}

		if strings.HasPrefix(sl.text, "## ") {
			break
		}
		result = append(result, sl)
	}
	return result
}

func (ps *parsedSpec) linesInBehavior(name string) []specLine {
	startLine, ok := ps.behaviorLines[name]
	if !ok {
		return nil
	}

	var result []specLine
	inSection := false
	fd := 0

	for _, sl := range ps.lines {
		if sl.num == startLine {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}

		trimmed := strings.TrimSpace(sl.text)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fd == 0 {
				fd = 1
			} else {
				fd--
			}
			continue
		}
		if fd > 0 {
			continue
		}

		if strings.HasPrefix(sl.text, "## ") {
			break
		}
		result = append(result, sl)
	}
	return result
}

func containsLine(lines []specLine, pred func(string) bool) (int, bool) {
	for _, sl := range lines {
		if pred(sl.text) {
			return sl.num, true
		}
	}
	return 0, false
}

// LintContent lints a PCD spec given as a string.
// filename is used for diagnostics only and must end in ".md".
func LintContent(content, filename string) LintResult {
	result := LintResult{File: filename}

	ps, err := parseSpecFromString(content)
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: SevError, Section: "structure", Line: 1, Rule: "RULE-01",
			Message: fmt.Sprintf("cannot parse content: %s", err),
		})
		result.Errors = 1
		return result
	}

	var diags []Diagnostic
	add := func(sev Severity, section string, line int, rule, msg string) {
		diags = append(diags, Diagnostic{Severity: sev, Section: section, Line: line, Rule: rule, Message: msg})
	}

	// RULE-01: required sections
	requiredSections := []string{
		"## META", "## TYPES", "## PRECONDITIONS",
		"## POSTCONDITIONS", "## INVARIANTS", "## EXAMPLES",
	}
	for _, s := range requiredSections {
		if _, ok := ps.sections[s]; !ok {
			sectionName := strings.TrimPrefix(s, "## ")
			add(SevError, "structure", 1, "RULE-01", fmt.Sprintf("Missing required section: %s", sectionName))
		}
	}
	if !ps.hasBehaviorSection() {
		add(SevError, "structure", 1, "RULE-01", "Missing required section: ## BEHAVIOR")
	}

	// RULE-02: required META fields
	requiredMeta := []string{"Deployment", "Verification", "Safety-Level", "Version", "Spec-Schema", "License"}
	for _, f := range requiredMeta {
		if _, ok := ps.metaFields[f]; !ok {
			lineNum := 1
			if ln, ok2 := ps.sections["## META"]; ok2 {
				lineNum = ln
			}
			add(SevError, "META", lineNum, "RULE-02", fmt.Sprintf("Missing required META field: %s", f))
		} else if ps.metaFields[f] == "" {
			ln := ps.metaLineNums[f]
			add(SevError, "META", ln, "RULE-02", fmt.Sprintf("META field %s has empty value", f))
		}
	}

	// RULE-02b: Author required
	if len(ps.metaAuthors) == 0 {
		metaLine := 1
		if ln, ok := ps.sections["## META"]; ok {
			metaLine = ln
		}
		add(SevError, "META", metaLine, "RULE-02", "Missing required META field: Author (at least one Author: line required)")
	} else {
		for _, a := range ps.metaAuthors {
			if a == "" {
				add(SevError, "META", ps.authorLineNum, "RULE-02", "Author: field has empty value")
			}
		}
	}

	// RULE-02c: Version semver
	if v, ok := ps.metaFields["Version"]; ok && v != "" {
		if !reSemanticVersion.MatchString(v) {
			ln := ps.metaLineNums["Version"]
			add(SevError, "META", ln, "RULE-02", fmt.Sprintf(
				"Version '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", v))
		}
	}

	// RULE-02d: Spec-Schema semver
	if s, ok := ps.metaFields["Spec-Schema"]; ok && s != "" {
		if !reSemanticVersion.MatchString(s) {
			ln := ps.metaLineNums["Spec-Schema"]
			add(SevError, "META", ln, "RULE-02", fmt.Sprintf(
				"Spec-Schema '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", s))
		}
	}

	// RULE-02e: SPDX license
	if l, ok := ps.metaFields["License"]; ok && l != "" {
		if !IsValidSPDX(l) {
			ln := ps.metaLineNums["License"]
			add(SevError, "META", ln, "RULE-02", fmt.Sprintf(
				"License '%s' is not a valid SPDX identifier. See https://spdx.org/licenses/ for valid identifiers.", l))
		}
	}

	// RULE-03: known deployment template
	deployment := ps.metaFields["Deployment"]
	if deployment != "" {
		if deployment == "crypto-library" {
			add(SevError, "META", 1, "RULE-03",
				"Deployment 'crypto-library' was retired in 0.3.6. Use 'verified-library' instead.")
		}

		known := false
		for _, t := range KnownTemplates {
			if t == deployment {
				known = true
				break
			}
		}
		if !known && deployment != "crypto-library" {
			ln := ps.metaLineNums["Deployment"]
			add(SevError, "META", ln, "RULE-03", fmt.Sprintf(
				"Unknown deployment template: '%s'. Run 'pcd-lint list-templates' to see valid values.", deployment))
		}

		if deployment == "enhance-existing" {
			lang, hasLang := ps.metaFields["Language"]
			if !hasLang {
				ln := ps.metaLineNums["Deployment"]
				add(SevError, "META", ln, "RULE-03", "Deployment 'enhance-existing' requires META field 'Language'")
			} else if lang == "" {
				ln := ps.metaLineNums["Language"]
				add(SevError, "META", ln, "RULE-03", "META field 'Language' has empty value")
			}
		}

		if deployment == "manual" {
			if _, hasTarget := ps.metaFields["Target"]; !hasTarget {
				ln := ps.metaLineNums["Deployment"]
				add(SevError, "META", ln, "RULE-03",
					"Deployment 'manual' requires META field 'Target'")
			}
		}

		if deployment == "python-tool" {
			sl := ps.metaFields["Safety-Level"]
			if sl != "QM" {
				ln := ps.metaLineNums["Safety-Level"]
				add(SevError, "META", ln, "RULE-03",
					"Deployment 'python-tool' requires Safety-Level: QM.")
			}
			vf := ps.metaFields["Verification"]
			if vf != "none" {
				ln := ps.metaLineNums["Verification"]
				add(SevError, "META", ln, "RULE-03",
					"Deployment 'python-tool' requires Verification: none.")
			}
		}

		if deployment == "verified-library" {
			sl := ps.metaFields["Safety-Level"]
			if sl == "QM" {
				ln := ps.metaLineNums["Safety-Level"]
				add(SevWarning, "META", ln, "RULE-03",
					"Deployment 'verified-library' with Safety-Level: QM is unusual.")
			}
		}
	}

	// RULE-04: deprecated META fields
	if _, hasTarget := ps.metaFields["Target"]; hasTarget && deployment != "manual" {
		ln := ps.metaLineNums["Target"]
		add(SevWarning, "META", ln, "RULE-04",
			"META field 'Target' is deprecated since v0.3.0.")
	}
	if _, hasDomain := ps.metaFields["Domain"]; hasDomain {
		ln := ps.metaLineNums["Domain"]
		add(SevWarning, "META", ln, "RULE-04",
			"META field 'Domain' is deprecated since v0.3.0. Use 'Deployment' instead.")
	}

	// RULE-05: known Verification values
	knownVerif := map[string]bool{"none": true, "lean4": true, "fstar": true, "dafny": true, "custom": true}
	if v, ok := ps.metaFields["Verification"]; ok && v != "" {
		if !knownVerif[v] {
			ln := ps.metaLineNums["Verification"]
			add(SevWarning, "META", ln, "RULE-05", fmt.Sprintf(
				"Unknown verification value: '%s'. Known values: none, lean4, fstar, dafny, custom.", v))
		}
	}

	// RULE-06 & RULE-07: example blocks
	applyExampleRules(ps, add)

	// RULE-08: BEHAVIOR must have STEPS
	for _, name := range ps.behaviorNames {
		blines := ps.linesInBehavior(name)
		_, hasSteps := containsLine(blines, func(s string) bool {
			return strings.HasPrefix(s, "STEPS:")
		})
		if !hasSteps {
			ln := ps.behaviorLines[name]
			add(SevError, fmt.Sprintf("BEHAVIOR: %s", name), ln, "RULE-08",
				fmt.Sprintf("BEHAVIOR '%s' is missing required STEPS: block.", name))
		}
	}

	// RULE-09: INVARIANTS must have tags
	if _, ok := ps.sections["## INVARIANTS"]; ok {
		invLines := ps.linesInSection("## INVARIANTS")
		for _, sl := range invLines {
			t := strings.TrimSpace(sl.text)
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			if strings.HasPrefix(sl.text, "- ") {
				if !strings.HasPrefix(sl.text, "- [observable]") && !strings.HasPrefix(sl.text, "- [implementation]") {
					add(SevWarning, "INVARIANTS", sl.num, "RULE-09",
						"Invariant entry missing tag. Prefix with [observable] or [implementation].")
				}
			}
		}
	}

	// RULE-10: error paths need negative examples
	applyRule10(ps, add)

	// RULE-11: TOOLCHAIN-CONSTRAINTS constraint values
	if _, ok := ps.sections["## TOOLCHAIN-CONSTRAINTS"]; ok {
		tcLines := ps.linesInSection("## TOOLCHAIN-CONSTRAINTS")
		for _, sl := range tcLines {
			t := strings.TrimSpace(sl.text)
			if t == "" {
				continue
			}
			if strings.Contains(t, "required") || strings.Contains(t, "forbidden") {
				continue
			}
			if strings.HasPrefix(t, "-") || strings.Contains(t, ":") {
				add(SevWarning, "TOOLCHAIN-CONSTRAINTS", sl.num, "RULE-11",
					"TOOLCHAIN-CONSTRAINTS entry uses unknown constraint value. Valid values: required, forbidden.")
			}
		}
	}

	// RULE-12: types not redeclared in BEHAVIOR
	applyRule12(ps, add)

	// RULE-13: valid Constraint values in BEHAVIOR
	validConstraints := map[string]bool{"required": true, "supported": true, "forbidden": true}
	for _, name := range ps.behaviorNames {
		blines := ps.linesInBehavior(name)
		for _, sl := range blines {
			if strings.HasPrefix(sl.text, "Constraint:") {
				val := strings.TrimSpace(strings.TrimPrefix(sl.text, "Constraint:"))
				if !validConstraints[val] {
					ln := ps.behaviorLines[name]
					add(SevError, fmt.Sprintf("BEHAVIOR: %s", name), ln, "RULE-13",
						fmt.Sprintf("BEHAVIOR '%s' has invalid Constraint: value '%s'. Valid values: required, supported, forbidden.", name, val))
				}
				if val == "forbidden" {
					_, hasReason := containsLine(blines, func(s string) bool {
						return strings.HasPrefix(s, "  reason:")
					})
					if !hasReason {
						ln := ps.behaviorLines[name]
						add(SevWarning, fmt.Sprintf("BEHAVIOR: %s", name), ln, "RULE-13",
							fmt.Sprintf("BEHAVIOR '%s' is Constraint: forbidden but has no reason: annotation.", name))
					}
				}
				break
			}
		}
	}

	// RULE-14: template EXECUTION section
	if deployment == "template" {
		if _, hasExec := ps.sections["## EXECUTION"]; !hasExec {
			add(SevWarning, "structure", 1, "RULE-14",
				"Deployment template is missing ## EXECUTION section.")
		} else {
			execLines := ps.linesInSection("## EXECUTION")
			execText := ""
			for _, sl := range execLines {
				execText += sl.text + "\n"
			}
			if !strings.Contains(execText, "### Delivery phases") {
				add(SevWarning, "EXECUTION", ps.sections["## EXECUTION"], "RULE-14",
					"## EXECUTION section has no '### Delivery phases' subsection.")
			}
			if !strings.Contains(execText, "### Compile gate") && !strings.Contains(execText, "COMPILE-GATE: none") {
				add(SevWarning, "EXECUTION", ps.sections["## EXECUTION"], "RULE-14",
					"## EXECUTION section has no '### Compile gate' subsection.")
			}
			if !strings.Contains(execText, "### Resume logic") {
				add(SevWarning, "EXECUTION", ps.sections["## EXECUTION"], "RULE-14",
					"## EXECUTION section has no '### Resume logic' subsection.")
			}
		}
	}

	// RULE-15, 16, 17: milestones
	if len(ps.milestones) > 0 {
		applyRule15(ps, add)
		applyRule16(ps, add)
		applyRule17(ps, add)
	}

	// Sort by line number
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Line < diags[j].Line
	})

	result.Diagnostics = diags

	for _, d := range diags {
		if d.Severity == SevError {
			result.Errors++
		} else {
			result.Warnings++
		}
	}
	result.Valid = result.Errors == 0

	return result
}

// ── RULE-06 / RULE-07 ─────────────────────────────────────────────────────────

type exampleBlock struct {
	name            string
	lineNum         int
	hasGiven        bool
	hasWhen         bool
	hasThen         bool
	whenWithoutThen bool
	givenEmpty      bool
	whenEmpty       bool
	thenEmpty       bool
}

func applyExampleRules(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	examplesLine, ok := ps.sections["## EXAMPLES"]
	if !ok {
		return
	}

	exLines := ps.linesInSection("## EXAMPLES")

	type state int
	const (
		stateNone  state = iota
		stateGiven
		stateWhen
		stateThen
	)

	var blocks []exampleBlock
	var cur *exampleBlock
	curState := stateNone
	givenHasContent := false
	whenHasContent := false
	thenHasContent := false
	pendingWhen := false

	finishBlock := func() {
		if cur == nil {
			return
		}
		if curState == stateThen && !thenHasContent {
			cur.thenEmpty = true
		}
		if curState == stateWhen && !whenHasContent {
			cur.whenEmpty = true
		}
		if pendingWhen {
			cur.whenWithoutThen = true
		}
		blocks = append(blocks, *cur)
		cur = nil
		curState = stateNone
		pendingWhen = false
	}

	for _, sl := range exLines {
		raw := sl.text

		if strings.HasPrefix(raw, "EXAMPLE:") {
			finishBlock()
			name := strings.TrimSpace(strings.TrimPrefix(raw, "EXAMPLE:"))
			cur = &exampleBlock{name: name, lineNum: sl.num}
			curState = stateNone
			givenHasContent = false
			whenHasContent = false
			thenHasContent = false
			pendingWhen = false
			continue
		}

		if cur == nil {
			continue
		}

		if strings.HasPrefix(raw, "GIVEN:") {
			if curState == stateWhen && !whenHasContent {
				cur.whenEmpty = true
			}
			if curState == stateThen && !thenHasContent {
				cur.thenEmpty = true
			}
			cur.hasGiven = true
			curState = stateGiven
			givenHasContent = false
			continue
		}

		if strings.HasPrefix(raw, "WHEN:") {
			if curState == stateGiven && !givenHasContent {
				cur.givenEmpty = true
			}
			if curState == stateThen && !thenHasContent {
				cur.thenEmpty = true
			}
			if pendingWhen {
				cur.whenWithoutThen = true
			}
			cur.hasWhen = true
			curState = stateWhen
			inlineContent := strings.TrimSpace(strings.TrimPrefix(raw, "WHEN:"))
			whenHasContent = inlineContent != ""
			pendingWhen = true
			continue
		}

		if strings.HasPrefix(raw, "THEN:") {
			if curState == stateWhen && !whenHasContent {
				cur.whenEmpty = true
			}
			cur.hasThen = true
			curState = stateThen
			thenHasContent = false
			pendingWhen = false
			continue
		}

		if strings.TrimSpace(raw) != "" {
			switch curState {
			case stateGiven:
				givenHasContent = true
			case stateWhen:
				whenHasContent = true
			case stateThen:
				thenHasContent = true
			}
		}
	}
	finishBlock()

	if len(blocks) == 0 {
		add(SevError, "EXAMPLES", examplesLine, "RULE-06",
			"EXAMPLES section contains no example blocks. Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers.")
		return
	}

	for _, b := range blocks {
		if !b.hasGiven {
			add(SevError, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' missing GIVEN: marker", b.name))
		}
		if !b.hasWhen {
			add(SevError, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' missing WHEN: marker", b.name))
		}
		if !b.hasThen {
			add(SevError, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' missing THEN: marker", b.name))
		}
		if b.whenWithoutThen {
			add(SevError, "EXAMPLES", b.lineNum, "RULE-07", fmt.Sprintf("Example '%s' has WHEN: without a matching THEN:", b.name))
		}
		if b.givenEmpty {
			add(SevWarning, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' has empty GIVEN block", b.name))
		}
		if b.whenEmpty {
			add(SevWarning, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' has empty WHEN block", b.name))
		}
		if b.thenEmpty {
			add(SevWarning, "EXAMPLES", b.lineNum, "RULE-06", fmt.Sprintf("Example '%s' has empty THEN block", b.name))
		}
	}
}

// ── RULE-10 ───────────────────────────────────────────────────────────────────

func applyRule10(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	exLines := ps.linesInSection("## EXAMPLES")

	negativePatterns := []string{
		"Err(", "error", "exit_code = 1", "exit_code = 2", "stderr contains",
		"exit 1", "exit 2", "errors contains", "errors non-empty",
	}

	for _, name := range ps.behaviorNames {
		blines := ps.linesInBehavior(name)
		hasErrorExit := false
		for _, sl := range blines {
			if strings.Contains(sl.text, "→") {
				hasErrorExit = true
				break
			}
		}
		if !hasErrorExit {
			continue
		}

		hasNegative := false
		inThen := false
		for _, sl := range exLines {
			raw := sl.text
			if strings.HasPrefix(raw, "THEN:") {
				inThen = true
				continue
			}
			if strings.HasPrefix(raw, "WHEN:") || strings.HasPrefix(raw, "GIVEN:") || strings.HasPrefix(raw, "EXAMPLE:") {
				inThen = false
				continue
			}
			if inThen {
				for _, p := range negativePatterns {
					if strings.Contains(strings.ToLower(sl.text), strings.ToLower(p)) {
						hasNegative = true
						break
					}
				}
			}
		}

		if !hasNegative {
			ln := ps.behaviorLines[name]
			add(SevError, fmt.Sprintf("BEHAVIOR: %s", name), ln, "RULE-10",
				fmt.Sprintf("BEHAVIOR '%s' has error exits in STEPS but no negative-path EXAMPLE.", name))
		}
	}
}

// ── RULE-12 ───────────────────────────────────────────────────────────────────

func applyRule12(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	typeNames := []string{}
	if _, ok := ps.sections["## TYPES"]; ok {
		typeLines := ps.linesInSection("## TYPES")
		for _, sl := range typeLines {
			if m := reTypeDecl.FindStringSubmatch(sl.text); m != nil {
				typeNames = append(typeNames, m[1])
			}
		}
	}

	for _, name := range ps.behaviorNames {
		blines := ps.linesInBehavior(name)
		for _, sl := range blines {
			for _, t := range typeNames {
				pattern := t + " :="
				if strings.Contains(sl.text, pattern) {
					ln := ps.behaviorLines[name]
					add(SevError, fmt.Sprintf("BEHAVIOR: %s", name), ln, "RULE-12",
						fmt.Sprintf("Type '%s' declared in TYPES is redefined in BEHAVIOR.", t))
				}
			}
		}
	}
}

// ── RULE-15 ───────────────────────────────────────────────────────────────────

func applyRule15(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	validStatuses := map[string]bool{"pending": true, "active": true, "failed": true, "released": true}

	for _, m := range ps.milestones {
		sec := fmt.Sprintf("MILESTONE: %s", m.name)

		if !m.hasIncluded {
			add(SevError, sec, m.lineNum, "RULE-15",
				fmt.Sprintf("MILESTONE '%s' is missing required 'Included BEHAVIORs:' field.", m.name))
		}

		isScaffold := m.scaffold == "true"
		if !m.hasDeferred && !isScaffold {
			add(SevError, sec, m.lineNum, "RULE-15",
				fmt.Sprintf("MILESTONE '%s' is missing required 'Deferred BEHAVIORs:' field.", m.name))
		}

		if !m.hasAcceptance {
			add(SevWarning, sec, m.lineNum, "RULE-15",
				fmt.Sprintf("MILESTONE '%s' has no 'Acceptance criteria:' field.", m.name))
		}

		if !m.hasStatus {
			add(SevWarning, sec, m.lineNum, "RULE-15",
				fmt.Sprintf("MILESTONE '%s' has no Status: field.", m.name))
		} else {
			if !validStatuses[m.status] {
				add(SevError, sec, m.statusLineNum, "RULE-15",
					fmt.Sprintf("MILESTONE '%s' has invalid Status: value '%s'.", m.name, m.status))
			}
		}

		if m.hasScaffold {
			if m.scaffold != "true" && m.scaffold != "false" {
				add(SevError, sec, m.scaffoldLineNum, "RULE-15",
					fmt.Sprintf("MILESTONE '%s' has invalid Scaffold: value '%s'.", m.name, m.scaffold))
			}
		}
	}

	activeCount := 0
	for _, m := range ps.milestones {
		if m.status == "active" {
			activeCount++
		}
	}
	if activeCount > 1 {
		add(SevError, "structure", 1, "RULE-15",
			"More than one MILESTONE has Status: active. Exactly one milestone may be active at a time.")
	}
}

// ── RULE-16 ───────────────────────────────────────────────────────────────────

func applyRule16(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	behaviorSet := make(map[string]bool)
	for _, n := range ps.behaviorNames {
		behaviorSet[n] = true
	}

	for _, m := range ps.milestones {
		sec := fmt.Sprintf("MILESTONE: %s", m.name)
		for _, n := range m.includedBehavs {
			if !behaviorSet[n] {
				add(SevError, sec, m.lineNum, "RULE-16",
					fmt.Sprintf("MILESTONE '%s' lists BEHAVIOR '%s' under Included BEHAVIORs but no such BEHAVIOR exists.", m.name, n))
			}
		}
		for _, n := range m.deferredBehavs {
			if !behaviorSet[n] {
				add(SevError, sec, m.lineNum, "RULE-16",
					fmt.Sprintf("MILESTONE '%s' lists BEHAVIOR '%s' under Deferred BEHAVIORs but no such BEHAVIOR exists.", m.name, n))
			}
		}
	}
}

// ── RULE-17 ───────────────────────────────────────────────────────────────────

func applyRule17(ps *parsedSpec, add func(Severity, string, int, string, string)) {
	var scaffolds []milestone
	for _, m := range ps.milestones {
		if m.scaffold == "true" {
			scaffolds = append(scaffolds, m)
		}
	}

	if len(scaffolds) > 1 {
		add(SevError, "structure", 1, "RULE-17",
			"more than one MILESTONE has Scaffold: true. At most one scaffold milestone is permitted per spec.")
	}

	if len(scaffolds) == 1 {
		sm := scaffolds[0]
		first := ps.milestones[0]
		if sm.name != first.name {
			add(SevError, fmt.Sprintf("MILESTONE: %s", sm.name), sm.lineNum, "RULE-17",
				fmt.Sprintf("Scaffold milestone '%s' must appear first in the spec (must appear first in document order).", sm.name))
		}
	}
}
