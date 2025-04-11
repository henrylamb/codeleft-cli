package report

import (
	"codeleft-cli/filter" // Assuming this path is correct
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReportNode represents a node (file or directory) in the report tree.
type ReportNode struct {
	Name           string
	Path           string                // Full path relative to root
	IsDir          bool
	Details        []filter.GradeDetails // Stores ALL GradeDetails for this file (if IsDir is false)
	Children       []*ReportNode         // Populated for directories (using pointers)
	Coverage       float64               // Calculated OVERALL coverage for this node
	CoverageOk     bool                  // Flag if overall coverage was calculable
	ToolCoverages  map[string]float64    // Coverage per tool (file's tool coverage OR directory's average coverage per tool)
	ToolCoverageOk map[string]bool       // Flag if coverage for a specific tool was calculable/present
}

// ReportViewData holds all data needed by the HTML template.
type ReportViewData struct {
	RootNodes       []*ReportNode      // Top-level files/dirs (using pointers)
	AllTools        []string           // Sorted list of unique tools found
	OverallAverages map[string]float64 // Average coverage per tool across ALL files
	TotalAverage    float64            // Overall average coverage across ALL files/tools
	ThresholdGrade  string             // The threshold grade used for calculations
}

// GenerateRepoHTMLReport generates the HTML report.
// Takes a slice of GradeDetail structs as input.
func GenerateRepoHTMLReport(gradeDetails []filter.GradeDetails, outputPath string, thresholdGrade string) error {
	if len(gradeDetails) == 0 {
		log.Println("Warning: No grade details provided to generate report.")
		// Optionally create an empty/minimal report or return an error
		// For now, let's proceed and it will likely generate an empty table
	}

	// 1. Group GradeDetails by FileName (path)
	groupedDetails := groupGradeDetailsByPath(gradeDetails)

	// 2. Build the ReportNode tree structure from the grouped paths.
	//    This step only creates the hierarchy, not coverages yet.
	rootNodes := buildReportTree(groupedDetails)

	// 3. Calculate coverages (file, directory averages) recursively,
	//    and collect global stats (tool names, sums for overall averages).
	toolSet := make(map[string]struct{})
	globalToolCoverageSums := make(map[string]float64) // Sum of coverage per tool across ALL FILES
	globalToolFileCounts := make(map[string]int)     // Files assessed per tool across ALL FILES

	for _, node := range rootNodes {
		calculateNodeCoverages(node, groupedDetails, thresholdGrade, toolSet, globalToolCoverageSums, globalToolFileCounts)
	}

	// Sort the tree alphabetically (dirs first) after calculations if needed
	sortReportNodes(rootNodes) // Apply sorting recursively

	// Convert tool set to a sorted slice.
	allTools := make([]string, 0, len(toolSet))
	for tool := range toolSet {
		allTools = append(allTools, tool)
	}
	sort.Strings(allTools)

	// 4. Calculate OVERALL report averages.
	overallAverages := make(map[string]float64)
	var totalCoverageSum float64
	var totalUniqueFilesWithCoverage int // Count unique files with valid coverage

	// Iterate through the *original* grouped data to get file-level data accurately
	processedFilesForTotalAvg := make(map[string]struct{}) // Track files counted

	for filePath, detailsList := range groupedDetails {
		if _, alreadyProcessed := processedFilesForTotalAvg[filePath]; alreadyProcessed {
			continue
		}

		var fileCoverageSum float64
		var fileToolCount int
		fileHasValidCoverage := false

		// Recalculate file's average coverage based *only* on its own tools
		processedToolsThisFile := make(map[string]struct{}) // Handle multiple entries for same tool if needed
		for _, detail := range detailsList {
			if _, toolDone := processedToolsThisFile[detail.Tool]; toolDone {
				continue // Skip if we already processed this tool for this file
			}
			if detail.Tool != "" && detail.Grade != "" {
				cov := calculateCoverage(detail.Grade, thresholdGrade)
				fileCoverageSum += cov
				fileToolCount++
				processedToolsThisFile[detail.Tool] = struct{}{}
				fileHasValidCoverage = true // Mark that this file contributes
			}
		}

		if fileHasValidCoverage && fileToolCount > 0 {
			fileAvg := fileCoverageSum / float64(fileToolCount)
			totalCoverageSum += fileAvg // Add the file's *average* coverage to the total sum
			totalUniqueFilesWithCoverage++
			processedFilesForTotalAvg[filePath] = struct{}{}
		}
	}

	// Calculate average per tool using globally collected sums/counts
	for _, tool := range allTools {
		sum := globalToolCoverageSums[tool]
		count := globalToolFileCounts[tool]
		if count > 0 {
			overallAverages[tool] = sum / float64(count)
		} else {
			overallAverages[tool] = 0 // Or potentially math.NaN()
		}
	}

	// Calculate final total average across all unique files with coverage
	var totalAverage float64
	if totalUniqueFilesWithCoverage > 0 {
		totalAverage = totalCoverageSum / float64(totalUniqueFilesWithCoverage)
	}

	// 5. Prepare data for the template.
	viewData := ReportViewData{
		RootNodes:       rootNodes,
		AllTools:        allTools,
		OverallAverages: overallAverages,
		TotalAverage:    totalAverage,
		ThresholdGrade:  thresholdGrade,
	}

	// 6. Parse and execute the template.
	tmpl, err := template.New("repoReport").Funcs(templateFuncs).Parse(repoReportTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

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

// groupGradeDetailsByPath groups the flat list of details into a map
// where the key is the file path (FileName) and the value is a slice
// of all GradeDetails for that path.
func groupGradeDetailsByPath(details []filter.GradeDetails) map[string][]filter.GradeDetails {
	grouped := make(map[string][]filter.GradeDetails)
	for _, d := range details {
		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(d.FileName)
		grouped[normalizedPath] = append(grouped[normalizedPath], d)
	}
	return grouped
}

// buildReportTree constructs the basic tree hierarchy from file paths.
// It does not calculate coverage here.
func buildReportTree(groupedDetails map[string][]filter.GradeDetails) []*ReportNode {
	roots := []*ReportNode{}
	// Use a map to keep track of created directory nodes by their full path
	// Ensures we don't create duplicate nodes for the same directory
	dirs := make(map[string]*ReportNode)

	// Sort paths for potentially more structured processing (optional but can help)
	paths := make([]string, 0, len(groupedDetails))
	for p := range groupedDetails {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, fullPath := range paths {
		details := groupedDetails[fullPath] // Get the details for this file
		parts := strings.Split(fullPath, "/")
		if len(parts) == 0 {
			continue // Skip empty paths
		}

		var parent *ReportNode
		currentPath := ""

		for i, part := range parts {
			isLastPart := (i == len(parts)-1)
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			// Check if node already exists (could be a dir created by a previous path)
			existingNode, found := dirs[currentPath]

			if isLastPart { // This is the file part
				fileNode := &ReportNode{
					Name:           part,
					Path:           fullPath, // Store the full original path
					IsDir:          false,
					Details:        details, // Store associated details
					ToolCoverages:  make(map[string]float64),
					ToolCoverageOk: make(map[string]bool),
				}
				if parent == nil { // File in root
					roots = append(roots, fileNode)
				} else {
					parent.Children = append(parent.Children, fileNode)
				}
				// Don't add files to the 'dirs' map
			} else { // This is a directory part
				if found {
					// Directory node already exists, just update parent pointer
					parent = existingNode
				} else {
					// Create new directory node
					dirNode := &ReportNode{
						Name:           part,
						Path:           currentPath, // Path up to this directory
						IsDir:          true,
						Children:       []*ReportNode{},
						ToolCoverages:  make(map[string]float64),
						ToolCoverageOk: make(map[string]bool),
					}
					dirs[currentPath] = dirNode // Add to map for lookup

					if parent == nil { // Directory in root
						roots = append(roots, dirNode)
					} else {
						// Check if child already exists in parent (can happen with sorting/processing order)
						childExists := false
						for _, child := range parent.Children {
							if child.Path == dirNode.Path {
								childExists = true
								break
							}
						}
						if !childExists {
							parent.Children = append(parent.Children, dirNode)
						}
					}
					parent = dirNode // This new dir becomes the parent for the next part
				}
			}
		}
	}
	return roots
}

// calculateNodeCoverages recursively calculates coverage for nodes and collects global stats.
// It modifies the node directly.
func calculateNodeCoverages(
	node *ReportNode,
	groupedDetails map[string][]filter.GradeDetails, // Pass this down if needed, or use node.Details
	thresholdGrade string,
	toolSet map[string]struct{},
	globalToolCoverageSums map[string]float64,
	globalToolFileCounts map[string]int,
) {
	if node == nil {
		return
	}

	if !node.IsDir {
		// --- Process File Node ---
		var fileOverallCoverageSum float64
		var fileToolCount int
		processedToolsThisFile := make(map[string]struct{}) // Ensure each tool contributes once per file

		// Use the details stored directly on the node now
		for _, detail := range node.Details {
			if detail.Tool == "" || detail.Grade == "" {
				continue // Skip if tool or grade is missing
			}
			if _, toolDone := processedToolsThisFile[detail.Tool]; toolDone {
				continue // Only count first entry for a tool for this specific file node calculation
			}

			coverage := calculateCoverage(detail.Grade, thresholdGrade)
			node.ToolCoverages[detail.Tool] = coverage
			node.ToolCoverageOk[detail.Tool] = true
			toolSet[detail.Tool] = struct{}{} // Add tool to global set

			// Add to global sums/counts *only once* per file/tool combo
			globalToolCoverageSums[detail.Tool] += coverage
			globalToolFileCounts[detail.Tool]++

			fileOverallCoverageSum += coverage
			fileToolCount++
			processedToolsThisFile[detail.Tool] = struct{}{}
		}

		// Calculate the file's overall average coverage
		if fileToolCount > 0 {
			node.Coverage = fileOverallCoverageSum / float64(fileToolCount)
			node.CoverageOk = true
		} else {
			node.CoverageOk = false // No tools/grades found for this file
		}

	} else {
		// --- Process Directory Node ---
		// Recurse first to calculate children coverages
		for _, child := range node.Children {
			calculateNodeCoverages(child, groupedDetails, thresholdGrade, toolSet, globalToolCoverageSums, globalToolFileCounts)
		}

		// Now calculate this directory's averages based on its children
		var dirOverallCoverageSum float64
		dirNodesWithOverallCoverage := 0
		dirToolCoverageSums := make(map[string]float64)
		dirToolCoverageCounts := make(map[string]int)

		for _, child := range node.Children {
			// Aggregate overall coverage for the directory average
			if child.CoverageOk {
				dirOverallCoverageSum += child.Coverage
				dirNodesWithOverallCoverage++
			}

			// Aggregate per-tool coverage for the directory average
			for tool, coverage := range child.ToolCoverages {
				if child.ToolCoverageOk[tool] { // Check if the child had valid coverage for this tool
					dirToolCoverageSums[tool] += coverage
					dirToolCoverageCounts[tool]++
					// toolSet is already populated by file processing or deeper recursion
				}
			}
		}

		// Calculate and set the directory's overall average coverage
		if dirNodesWithOverallCoverage > 0 {
			node.Coverage = dirOverallCoverageSum / float64(dirNodesWithOverallCoverage)
			node.CoverageOk = true
		} else {
			node.CoverageOk = false // No children with valid coverage
		}

		// Calculate and set the directory's per-tool average coverage
		for tool, sum := range dirToolCoverageSums {
			count := dirToolCoverageCounts[tool]
			if count > 0 {
				node.ToolCoverages[tool] = sum / float64(count)
				node.ToolCoverageOk[tool] = true
			}
			// No need for else, ToolCoverageOk map default is false
		}
	}
}

// sortReportNodes recursively sorts children nodes: directories first, then alphabetically.
func sortReportNodes(nodes []*ReportNode) {
	// Sort the current level
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir // true (directory) comes before false (file)
		}
		return nodes[i].Name < nodes[j].Name
	})

	// Recursively sort children of directories
	for _, node := range nodes {
		if node.IsDir && len(node.Children) > 0 {
			sortReportNodes(node.Children)
		}
	}
}

// --- Utility functions (getGradeIndex, calculateCoverage) remain the same ---
func getGradeIndex(grade string) int {
	gradeIndices := map[string]int{
		"A*": 5, "A": 4, "B": 3, "C": 2, "D": 1, "F": 0,
	}
	// Ensure comparison is case-insensitive
	index, ok := gradeIndices[strings.ToUpper(grade)]
	if !ok {
		log.Printf("Warning: Unrecognized grade '%s', treating as F (0)", grade)
		return 0 // Default to lowest index for unrecognized grades
	}
	return index
}
func calculateCoverage(grade, thresholdGrade string) float64 {
	gradeIndex := getGradeIndex(grade)
	thresholdIndex := getGradeIndex(thresholdGrade)

	// Logic matches the JS example and previous Go version
	if gradeIndex > thresholdIndex {
		return 120.0
	} else if gradeIndex == thresholdIndex {
		return 100.0
	} else if gradeIndex >= thresholdIndex-1 {
		return 70.0
	} else if gradeIndex >= thresholdIndex-2 {
		return 50.0
	} else if gradeIndex >= thresholdIndex-3 {
		return 30.0
	} else {
		return 10.0
	}
}

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

// --- HTML Template (Removed Grade(s) and Tool(s) columns) ---
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