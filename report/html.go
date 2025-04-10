package report

import (
	"codeleft-cli/filter"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// --- Data Structures for Template ---

// ReportNode represents a node (file or directory) in the report tree.
type ReportNode struct {
	Name          string
	Path          string               // Full path relative to root for display/reference
	IsDir         bool
	Details       *filter.GradeDetails // Populated for files
	Children      []ReportNode         // Populated for directories
	Coverage      float64              // Calculated coverage for this node (file or avg of children)
	CoverageOk    bool                 // Flag if coverage was calculable (false if dir has no files below it)
	ToolCoverages map[string]float64   // Coverage per tool (relevant for averages/display)
}

// ReportViewData holds all data needed by the HTML template.
type ReportViewData struct {
	RootNodes       []ReportNode       // Top-level files/dirs
	AllTools        []string           // Sorted list of unique tools found
	OverallAverages map[string]float64 // Average coverage per tool across all files
	TotalAverage    float64            // Overall average coverage across all files/tools
	ThresholdGrade  string             // The threshold grade used for calculations
}

// --- Core Report Generation Function ---

// GenerateRepoHTMLReport creates an HTML report from the repo structure.
func GenerateRepoHTMLReport(repoStructure map[string]any, outputPath string, thresholdGrade string) error {
	// 1. Build the ReportNode tree and calculate file coverage.
	rootNodes := buildReportNodes(repoStructure, "", thresholdGrade)

	// 2. Calculate directory averages and collect all tool names.
	toolSet := make(map[string]struct{})
	toolFileCounts := make(map[string]int)     // Files per tool
	toolCoverageSums := make(map[string]float64) // Sum of coverage per tool

	calculateAveragesAndTools(rootNodes, toolSet, toolCoverageSums, toolFileCounts)

	// Convert tool set to a sorted slice.
	allTools := make([]string, 0, len(toolSet))
	for tool := range toolSet {
		allTools = append(allTools, tool)
	}
	sort.Strings(allTools)

	// 3. Calculate overall averages.
	overallAverages := make(map[string]float64)
	var totalCoverageSum float64
	var totalFileCount int

	for _, tool := range allTools {
		sum := toolCoverageSums[tool]
		count := toolFileCounts[tool]
		if count > 0 {
			overallAverages[tool] = sum / float64(count)
			totalCoverageSum += sum
			totalFileCount += count
		} else {
			overallAverages[tool] = 0
		}
	}

	var totalAverage float64
	if totalFileCount > 0 {
		totalAverage = totalCoverageSum / float64(totalFileCount)
	}

	// 4. Prepare data for the template.
	viewData := ReportViewData{
		RootNodes:       rootNodes,
		AllTools:        allTools,
		OverallAverages: overallAverages,
		TotalAverage:    totalAverage,
		ThresholdGrade:  thresholdGrade,
	}

	// 5. Parse and execute the template with our custom functions.
	tmpl, err := template.New("repoReport").Funcs(templateFuncs).Parse(repoReportTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Ensure output directory exists.
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML output file '%s': %w", outputPath, err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, viewData); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	fmt.Printf("Successfully generated repository report: %s\n", outputPath)
	return nil
}

// --- Helper Functions ---

// buildReportNodes recursively converts the map structure to a slice of ReportNodes.
func buildReportNodes(data map[string]any, currentPath string, thresholdGrade string) []ReportNode {
	nodes := make([]ReportNode, 0, len(data))
	for name, value := range data {
		nodePath := name
		if currentPath != "" {
			nodePath = fmt.Sprintf("%s/%s", currentPath, name)
		}

		node := ReportNode{
			Name: name,
			Path: nodePath,
		}

		switch v := value.(type) {
		case map[string]any: // Directory.
			node.IsDir = true
			node.Children = buildReportNodes(v, nodePath, thresholdGrade)
		case filter.GradeDetails: // File.
			node.IsDir = false
			details := v
			node.Details = &details
			node.Coverage = calculateCoverage(details.Grade, thresholdGrade)
			node.CoverageOk = true
			node.ToolCoverages = map[string]float64{details.Tool: node.Coverage}
		default:
			log.Printf("Warning: Unexpected type in repoStructure at path '%s': %T", nodePath, value)
			continue
		}
		nodes = append(nodes, node)
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir // Directories come before files.
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes
}

// calculateAveragesAndTools recursively calculates directory averages and collects tool info.
func calculateAveragesAndTools(nodes []ReportNode, toolSet map[string]struct{}, toolCoverageSums map[string]float64, toolFileCounts map[string]int) (float64, bool) {
	var totalCoverageSum float64
	var nodesWithCoverage int

	for i := range nodes {
		node := &nodes[i]
		if node.IsDir {
			dirAvg, dirOk := calculateAveragesAndTools(node.Children, toolSet, toolCoverageSums, toolFileCounts)
			node.Coverage = dirAvg
			node.CoverageOk = dirOk
			if dirOk {
				totalCoverageSum += dirAvg
				nodesWithCoverage++
			}
		} else {
			if node.Details != nil && node.Details.Tool != "" {
				toolSet[node.Details.Tool] = struct{}{}
				if node.CoverageOk {
					toolCoverageSums[node.Details.Tool] += node.Coverage
					toolFileCounts[node.Details.Tool]++
					totalCoverageSum += node.Coverage
					nodesWithCoverage++
				}
			}
		}
	}

	if nodesWithCoverage > 0 {
		return totalCoverageSum / float64(nodesWithCoverage), true
	}
	return 0, false
}

// getGradeIndex maps a grade string to a numeric value.
func getGradeIndex(grade string) int {
	gradeIndices := map[string]int{
		"A*": 5, "A": 4, "B": 3, "C": 2, "D": 1, "F": 0,
	}
	index, ok := gradeIndices[strings.ToUpper(grade)]
	if !ok {
		return 0
	}
	return index
}

// calculateCoverage computes the coverage percentage given a grade and threshold.
func calculateCoverage(grade, thresholdGrade string) float64 {
	gradeIndex := getGradeIndex(grade)
	thresholdIndex := getGradeIndex(thresholdGrade)

	switch {
	case gradeIndex > thresholdIndex:
		return 120
	case gradeIndex == thresholdIndex:
		return 100
	case gradeIndex >= thresholdIndex-1:
		return 70
	case gradeIndex >= thresholdIndex-2:
		return 50
	case gradeIndex >= thresholdIndex-3:
		return 30
	default:
		return 10
	}
}

// --- Template Functions and Helpers ---

// split is a helper function that splits a string by a given separator.
func split(s string, sep string) []string {
	return strings.Split(s, sep)
}

// dict constructs a map from a variadic list of key/value pairs.
// Keys must be strings.
func dict(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("dict function requires an even number of arguments")
	}
	d := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("dict keys must be strings")
		}
		d[key] = values[i+1]
	}
	return d
}

var templateFuncs = template.FuncMap{
	// Formatting and coverage helpers.
	"formatFloat": func(f float64) string {
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return fmt.Sprintf("%.2f", f)
		}
		return "N/a"
	},
	"getCoverageClass": func(coverage float64) string {
		// Keep class names semantic, colors are handled by CSS
		if coverage > 80 {
			return "green" // Represents good coverage
		} else if coverage >= 50 {
			return "orange" // Represents medium coverage
		}
		return "red" // Represents low coverage
	},
	"getCoverageColor": func(coverage float64) string {
		// These colors should work reasonably well on a dark background
		if coverage > 80 {
			return "#76C474" // Green
		} else if coverage >= 50 {
			return "#F0AB86" // Orange
		}
		return "rgb(224, 66, 66)" // Red
	},
	"getToolAverage": func(averages map[string]float64, tool string) float64 {
		if avg, ok := averages[tool]; ok {
			return avg
		}
		return 0
	},
	"split": split,

	// Arithmetic functions.
	"multiply": func(a, b int) int { return a * b },
	"sub":      func(a, b int) int { return a - b },
	"add":      func(a, b int) int { return a + b },
	"plus":     func(a, b int) int { return a + b },

	// Sequence and slice helper functions.
	"seq": func(n int) []int {
		s := make([]int, n)
		for i := 0; i < n; i++ {
			s[i] = i
		}
		return s
	},
	"makeSlice": func() []int {
		return []int{}
	},
	"append": func(slice []int, elem int) []int {
		return append(slice, elem)
	},
	"seqSimple": func(start, end int) []int {
		s := []int{}
		for i := start; i < end; i++ {
			s = append(s, i)
		}
		return s
	},
	"loop": func(n int) []int {
		s := make([]int, n)
		for i := 0; i < n; i++ {
			s[i] = i
		}
		return s
	},

	// Dictionary helper.
	"dict": dict,
}

// --- HTML Template (Dark Theme) ---
// We pass a dictionary with two keys: "Nodes" and "Root".
// "Nodes" is a slice of ReportNode to be rendered.
// "Root" is the full ReportViewData to preserve global context.
const repoReportTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Repository Structure Report</title>
    <meta charset="UTF-8">
    <style>
        /* --- Dark Theme CSS --- */
        :root {
            --bg-color: #1e1e1e;         /* Dark background */
            --text-color: #e0e0e0;       /* Light text */
            --border-color: #444444;     /* Darker border */
            --header-bg-color: #2a2a2a;  /* Slightly lighter dark bg for headers */
            --green-color: #76C474;      /* Green (kept, visible on dark) */
            --orange-color: #F0AB86;     /* Orange (kept, visible on dark) */
            --red-color: #e04242;        /* Slightly adjusted Red for visibility */
            --blue-color: #58a6ff;       /* Lighter blue for links/folders */
            --grey-color: #aaaaaa;       /* Lighter grey for secondary text/icons */
            --light-red-bg: rgba(224, 66, 66, 0.2); /* Darker translucent red bg */
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            line-height: 1.5;
            padding: 20px;
            background-color: var(--bg-color);
            color: var(--text-color);
            font-size: 14px;
        }
        h1, h2 {
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 0.3em;
            margin-bottom: 1em;
            margin-top: 1.5em;
            color: #cccccc; /* Slightly lighter heading color */
        }
        h1 { font-size: 1.8em; font-weight: 600; }
        h2 { font-size: 1.5em; font-weight: 600; }
        .summary {
            margin-bottom: 20px;
            font-size: 1.1em;
            display: flex;
            justify-content: space-between;
            flex-wrap: wrap;
            background-color: var(--header-bg-color);
            padding: 15px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
        }
        .summary div { margin: 5px 10px; }
        table {
            width: 100%;
            border-collapse: collapse;
            text-align: left;
            margin-bottom: 30px;
            border: 1px solid var(--border-color);
            border-radius: 6px;
            overflow: hidden;
        }
        th, td {
            padding: 10px 12px;
            border-bottom: 1px solid var(--border-color);
            vertical-align: middle;
        }
        th {
            background-color: var(--header-bg-color);
            font-weight: 600;
            text-align: center;
            color: #f0f0f0; /* Ensure header text is light */
        }
        td:first-child, th:first-child { text-align: left; }
        /* Remove last border in tbody and thead */
        tbody tr:last-child td { border-bottom: none; }
        thead tr:last-child th { border-bottom: 1px solid var(--border-color); } /* Keep border below header */

        /* Style rows for better readability on dark theme */
        tbody tr:nth-child(even) {
             background-color: rgba(255, 255, 255, 0.03); /* Subtle alternating row color */
        }
        tbody tr:hover {
            background-color: rgba(255, 255, 255, 0.06); /* Subtle hover effect */
        }

        .coverage-cell {
            text-align: center;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100%;
        }
        .coverage-text { font-weight: 600; font-size: 0.95em; }
        /* Coverage color classes - use color variables */
        .green { color: var(--green-color); }
        .orange { color: var(--orange-color); }
        .red { color: var(--red-color); }
        .grey { color: var(--grey-color); text-align: center; }
        .progress-bar {
            width: 80px;
            height: 8px;
            background-color: #555555; /* Darker background for progress bar */
            border-radius: 4px;
            overflow: hidden;
            margin-top: 4px;
        }
        .progress-fill {
            height: 100%;
            transition: width 0.3s ease;
            border-radius: 4px;
        }
        .low-coverage { background-color: var(--light-red-bg); }
        /* Ensure low-coverage hover is distinct */
        .low-coverage:hover {
             background-color: rgba(224, 66, 66, 0.3);
        }
        .folder-name { font-weight: 600; color: var(--blue-color); }
        .file-name {
           font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace;
           font-size: 0.9em;
           color: var(--text-color); /* Ensure file names use main text color */
        }
        .icon {
            display: inline-block;
            vertical-align: text-bottom;
            margin-right: 5px;
            fill: currentColor; /* Icons inherit color */
        }
        /* Icon colors now directly use the variables */
        .icon-folder { color: var(--blue-color); }
        .icon-file { color: var(--grey-color); }
        .grade-cell { text-align: center; font-weight: bold; }
        .tool-cell { text-align: center; font-size: 0.9em; color: var(--grey-color); }
    </style>
</head>
<body>
    <h1>Repository Structure Report</h1>
    <div class="summary">
        <div>Threshold Grade: <strong>{{ .ThresholdGrade }}</strong></div>
        <div>Overall Coverage:
            <strong class="coverage-text {{ getCoverageClass .TotalAverage }}">
                {{ formatFloat .TotalAverage }}%
            </strong>
        </div>
    </div>
    <h2>Detailed Coverage</h2>
    <table>
        <thead>
            <tr>
                <th>File / Directory</th>
                {{ range .AllTools }}
                    <th>{{ . }}</th>
                {{ end }}
                <th>Grade</th>
                <th>Tool</th>
                <th>Coverage (%)</th>
            </tr>
            <tr>
                <td><strong>Overall Averages</strong></td>
                {{ range .AllTools }}
                    {{ $avg := getToolAverage $.OverallAverages . }}
                    <td>
                        {{ if gt $avg 0.0 }}
                        <div class="coverage-cell">
                            <span class="coverage-text {{ getCoverageClass $avg }}">{{ formatFloat $avg }}%</span>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: {{ $avg }}%; background-color: {{ getCoverageColor $avg }};"></div>
                            </div>
                        </div>
                        {{ else }}
                        <span class="grey">N/a</span>
                        {{ end }}
                    </td>
                {{ end }}
                <td class="grey">N/a</td>
                <td class="grey">N/a</td>
                <td>
                    <div class="coverage-cell">
                        <span class="coverage-text {{ getCoverageClass .TotalAverage }}">{{ formatFloat .TotalAverage }}%</span>
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: {{ .TotalAverage }}%; background-color: {{ getCoverageColor .TotalAverage }};"></div>
                        </div>
                    </div>
                </td>
            </tr>
        </thead>
        <tbody>
            {{/* Start recursive rendering: pass a dict with "Nodes" and the full "Root" view */}}
            {{ template "nodeList" (dict "Nodes" .RootNodes "Root" .) }}
        </tbody>
    </table>

    {{/* --- Template Definitions --- */}}

    {{ define "nodeList" }}
        {{ $root := .Root }}
        {{ range .Nodes }}
            {{ template "node" (dict "Node" . "Root" $root) }}
        {{ end }}
    {{ end }}

    {{ define "node" }}
        {{ $node := .Node }}
        {{ $root := .Root }}
        <tr {{ if and (not $node.IsDir) (lt $node.Coverage 50.0) }} class="low-coverage" {{ end }}>
            <td>
                <span style="padding-left: {{ if gt (len (split $node.Path "/")) 1 }}{{ printf "%dpx" (multiply (sub (len (split $node.Path "/")) 1) 15) }}{{ else }}0px{{ end }};">
                    {{ if $node.IsDir }}
                        <svg class="icon icon-folder" width="16" height="16" viewBox="0 0 16 16" version="1.1">
                            <path fill-rule="evenodd" d="M1.75 1A1.75 1.75 0 000 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0016 13.25v-8.5A1.75 1.75 0 0014.25 3h-6.5a.25.25 0 01-.2-.1l-.9-1.2c-.33-.44-.85-.7-1.4-.7h-3.5z"></path>
                        </svg>
                        <span class="folder-name">{{ $node.Name }}</span>
                    {{ else }}
                        <svg class="icon icon-file" width="16" height="16" viewBox="0 0 16 16" version="1.1">
                            <path fill-rule="evenodd" d="M3.75 1.5a.25.25 0 01.25-.25h8.5a.25.25 0 01.25.25v13.25a.25.25 0 01-.25.25h-8.5a.25.25 0 01-.25-.25V1.5zm.5 0v13h7.5v-13h-7.5z"></path>
                        </svg>
                        <span class="file-name">{{ $node.Name }}</span>
                    {{ end }}
                </span>
            </td>
            {{ range $root.AllTools }}
                <td>
                    {{ if not $node.IsDir }}
                        {{ if eq $node.Details.Tool . }}
                            {{ $coverage := $node.Coverage }}
                            <div class="coverage-cell">
                                <span class="coverage-text {{ getCoverageClass $coverage }}">{{ formatFloat $coverage }}%</span>
                            </div>
                        {{ else }}
                            <span class="grey">-</span>
                        {{ end }}
                    {{ else }}
                       {{/* Empty cell for directory rows in tool columns */}}
                       <span class="grey"></span>
                    {{ end }}
                </td>
            {{ end }}
            <td class="grade-cell">
                {{ if not $node.IsDir }}{{ $node.Details.Grade }}{{ else }}<span class="grey"></span>{{ end }}
            </td>
            <td class="tool-cell">
                {{ if not $node.IsDir }}{{ $node.Details.Tool }}{{ else }}<span class="grey"></span>{{ end }}
            </td>
            <td>
                {{ if $node.CoverageOk }}
                    <div class="coverage-cell">
                        <span class="coverage-text {{ getCoverageClass $node.Coverage }}">{{ formatFloat $node.Coverage }}%</span>
                        <div class="progress-bar">
                            <div class="progress-fill" style="width: {{ $node.Coverage }}%; background-color: {{ getCoverageColor $node.Coverage }};"></div>
                        </div>
                    </div>
                {{ else }}
                    <span class="grey">N/a</span>
                {{ end }}
            </td>
        </tr>
        {{ if $node.IsDir }}
            {{ template "nodeList" (dict "Nodes" $node.Children "Root" $root) }}
        {{ end }}
    {{ end }}
</body>
</html>
`