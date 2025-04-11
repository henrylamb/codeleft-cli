package report

import (
	"fmt"
	"html/template"
	"math"
	"path/filepath"
	"strings"
)

// --- Template Functions (Removed getFileGrade and getFileTool) ---
var templateFuncs = template.FuncMap{
	"formatFloat": func(f float64) string {
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return fmt.Sprintf("%.2f", f)
		}
		return "N/a"
	},
	"getCoverageClass": func(coverage float64) string {
		if coverage >= 100 { return "green" }
		if coverage >= 70 { return "green-med" }
		if coverage >= 50 { return "orange" }
		if coverage >= 30 { return "orange-low" }
		return "red"
	},
	"getCoverageColor": func(coverage float64) string {
		if coverage >= 100 { return "#76C474" }
		if coverage >= 70 { return "#a0d080" }
		if coverage >= 50 { return "#F0AB86" }
		if coverage >= 30 { return "#f5be9f" }
		return "#e04242"
	},
	"getToolAverage": func(averages map[string]float64, tool string) float64 {
		if avg, ok := averages[tool]; ok {
			return avg
		}
		return 0
	},
	// Modified to accept *ReportNode
	"getToolCoverage": func(node *ReportNode, tool string) float64 {
		if node == nil { return 0 }
		if cov, ok := node.ToolCoverages[tool]; ok && node.ToolCoverageOk[tool] {
			return cov
		}
		return 0 // Return 0 if not valid or doesn't exist
	},
	// Modified to accept *ReportNode
	"hasToolCoverage": func(node *ReportNode, tool string) bool {
		if node == nil { return false }
		exists, ok := node.ToolCoverageOk[tool]
		return ok && exists
	},
	// getFileGrade removed
	// getFileTool removed
	"split": func(s string, sep string) []string {
		return strings.Split(s, sep)
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
		}
		d := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			d[key] = values[i+1]
		}
		return d, nil
	},
	"multiply": func(a, b int) int { return a * b },
	"sub":      func(a, b int) int { return a - b },
	"add":      func(a, b int) int { return a + b },
	"base": func(p string) string {
		return filepath.Base(p)
	},
	"dirLevel": func(p string) int {
		p = filepath.ToSlash(p)
		count := strings.Count(p, "/")
		parts := strings.Split(p, "/")
		if len(parts) > 0 && parts[len(parts)-1] != "" {
			return len(parts) - 1
		}
		return count
	},
}

const repoReportTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Repository Structure Report</title>
    <meta charset="UTF-8">
    <style>
        /* --- Dark Theme CSS (Hardcoded Colors) --- */
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            line-height: 1.5;
            padding: 20px;
            background-color: #1e1e1e; /* Dark background */
            color: #e0e0e0;       /* Light text */
            font-size: 14px;
        }
        h1, h2 {
            border-bottom: 1px solid #444444; /* Darker border */
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
            background-color: #2a2a2a;  /* Slightly lighter dark bg */
            padding: 15px;
            border: 1px solid #444444; /* Darker border */
            border-radius: 6px;
        }
        .summary div { margin: 5px 10px; }
        table {
            width: 100%;
            border-collapse: collapse;
            text-align: left;
            margin-bottom: 30px;
            border: 1px solid #444444; /* Darker border */
            border-radius: 6px;
            overflow: hidden; /* Ensures border radius applies to content */
        }
        th, td {
            padding: 8px 10px; /* Slightly reduced padding */
            border-bottom: 1px solid #444444; /* Darker border */
            vertical-align: middle;
            text-align: center; /* Default center align for most cells */
            white-space: nowrap; /* Prevent tool names wrapping */
        }
        th {
            background-color: #2a2a2a; /* Slightly lighter dark bg */
            font-weight: 600;
            color: #f0f0f0; /* Ensure header text is light */
            position: sticky; /* Make header sticky */
            top: 0;
            z-index: 1;
        }
        /* Left-align the first column (Name) */
        td:first-child, th:first-child {
            text-align: left;
            white-space: normal; /* Allow file/dir names to wrap if needed */
         }
        tbody tr:last-child td { border-bottom: none; } /* Remove border for last row in tbody */
        thead tr:last-child th { border-bottom: 2px solid #555555; } /* Stronger border below header */

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
            min-width: 80px; /* Ensure cells have minimum width */
        }
        .coverage-text { font-weight: 600; font-size: 0.9em; } /* Slightly smaller */

        /* Coverage color classes */
        .green { color: #76C474; }       /* Green */
        .green-med { color: #a0d080; }   /* Lighter Green */
        .orange { color: #F0AB86; }      /* Orange */
        .orange-low { color: #f5be9f; }  /* Lighter Orange */
        .red { color: #e04242; }         /* Red */
        .grey { color: #888888; text-align: center; font-style: italic; } /* Adjusted grey */

        .progress-bar {
            width: 60px; /* Reduced width */
            height: 6px; /* Reduced height */
            background-color: #555555; /* Darker background for progress bar */
            border-radius: 3px; /* Adjusted radius */
            overflow: hidden;
            margin-top: 3px; /* Reduced margin */
        }
        .progress-fill {
            height: 100%;
            transition: width 0.3s ease;
            border-radius: 3px;
        }
        .folder-name { font-weight: 600; color: #58a6ff; } /* Lighter blue */
        .file-name {
           font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace;
           font-size: 0.9em;
           color: #c0c0c0; /* Slightly adjusted file name color */
        }
        .icon {
            display: inline-block;
            vertical-align: -2px; /* Align icon better with text */
            margin-right: 6px;
            fill: currentColor; /* Icons inherit color */
        }
        .icon-folder { color: #58a6ff; } /* Lighter blue */
        .icon-file { color: #999999; } /* Adjusted grey */
        /* .grade-cell and .tool-cell classes can be removed from CSS if desired */
    </style>
</head>
<body>
    <h1>Repository Structure Report</h1>
    <div class="summary">
        <div>Threshold Grade Used for Calculation: <strong>{{ .ThresholdGrade }}</strong></div>
        <div>Overall Report Coverage:
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
                {{/* Tool Headers */}}
                {{ range .AllTools }}
                    <th>{{ . }}</th>
                {{ end }}
                {{/* REMOVED: <th>Grade(s)</th> */}}
                {{/* REMOVED: <th>Tool(s)</th> */}}
                <th>Overall Coverage</th> {{/* Node's overall coverage */}}
            </tr>
            {{/* Overall Averages Row */}}
            <tr>
                <td><strong>Overall Report Averages</strong></td>
                {{ range .AllTools }}
                    {{ $avg := getToolAverage $.OverallAverages . }}
                    <td>
                        {{ if gt $avg 0.0 }} {{/* Only show if average is calculated */}}
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
                {{/* REMOVED: <td class="grey">-</td> Placeholder for Grade */}}
                {{/* REMOVED: <td class="grey">-</td> Placeholder for Tool(s) */}}
                <td> {{/* Overall Total Average */}}
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
            {{/* Start recursive rendering: pass Nodes and the full Root data */}}
            {{ template "nodeList" (dict "Nodes" .RootNodes "Root" . "Level" 0) }}
        </tbody>
    </table>

    {{/* --- Template Definitions --- */}}

    {{/* Define nodeList to handle recursion and pass indentation level */}}
    {{ define "nodeList" }}
        {{ $root := .Root }}
        {{ $level := .Level }}
        {{ range .Nodes }}
            {{ template "node" (dict "Node" . "Root" $root "Level" $level) }}
        {{ end }}
    {{ end }}

    {{/* Define node for rendering a single file or directory row */}}
    {{ define "node" }}
        {{ $node := .Node }} {{/* Node is now a *ReportNode */}}
        {{ $root := .Root }}
        {{ $level := .Level }}
         <tr>
            {{/* Column 1: Name with Indentation */}}
             {{ $nodePath := $node.Path }} {{/* Use the node's stored path */}}
             {{ $displayLevel := dirLevel $nodePath }}
             <td style="text-align: left; padding-left: {{ add (multiply $displayLevel 20) 10 }}px;">
                 {{ if $node.IsDir }}
                    <svg class="icon icon-folder" width="16" height="16" viewBox="0 0 16 16" version="1.1">
                        <path fill-rule="evenodd" d="M1.75 1A1.75 1.75 0 000 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0016 13.25v-8.5A1.75 1.75 0 0014.25 3h-6.5a.25.25 0 01-.2-.1l-.9-1.2c-.33-.44-.85-.7-1.4-.7h-3.5z"></path>
                    </svg>
                    <span class="folder-name">{{ $node.Name }}</span>
                {{ else }}
                    <svg class="icon icon-file" width="16" height="16" viewBox="0 0 16 16" version="1.1">
                         <path fill-rule="evenodd" d="M3.75 1.5a.25.25 0 01.25-.25h8.5a.25.25 0 01.25.25v13.25a.25.25 0 01-.25.25H4a.25.25 0 01-.25-.25V1.5zM4 1.75v13h7.5V1.75H4z"></path>
                     </svg>
                    <span class="file-name">{{ $node.Name }}</span>
                {{ end }}
            </td>

            {{/* Columns for Each Tool */}}
            {{ range $root.AllTools }}
                {{ $toolName := . }}
                <td>
                    {{/* Use pointer access for helper functions */}}
                    {{ if hasToolCoverage $node $toolName }}
                        {{ $toolCov := getToolCoverage $node $toolName }}
                        <div class="coverage-cell">
                            <span class="coverage-text {{ getCoverageClass $toolCov }}">{{ formatFloat $toolCov }}%</span>
                             <div class="progress-bar">
                                <div class="progress-fill" style="width: {{ $toolCov }}%; background-color: {{ getCoverageColor $toolCov }};"></div>
                            </div>
                        </div>
                    {{ else }}
                        <span class="grey">-</span>
                    {{ end }}
                </td>
            {{ end }}

            {{/* REMOVED: Column: Grade(s) (Files Only) */}}
            {{/*
            <td class="grade-cell">
                {{ if not $node.IsDir }}{{ getFileGrade $node }}{{ else }}<span class="grey"></span>{{ end }}
            </td>
            */}}

            {{/* REMOVED: Column: Tool(s) (Files Only) */}}

            {{/* Column: Overall Coverage for this Node */}}
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

        {{/* Recursive Call for Directory Children */}}
        {{ if $node.IsDir }}
            {{/* Pass node.Children directly as they are already pointers */}}
            {{ template "nodeList" (dict "Nodes" $node.Children "Root" $root "Level" (add $level 1)) }}
        {{ end }}
    {{ end }}

</body>
</html>
`