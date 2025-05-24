# codeleft-cli

**Version:** 1.0.8

## Overview

**codeleft-cli** is an open-source command-line tool that analyzes and assesses code quality based on user-defined thresholds and tooling preferences. This CLI is designed to:

1. **Automatically manage config files**: If no configuration or history file is found in the project root, it creates them with default data.
2. **Assess code quality**: It can optionally fail CI/CD pipelines if certain thresholds (e.g., grade or coverage) are not met.
3. **Filter by tooling**: Users can specify a set of code quality tools (e.g., SOLID, OWASP-Top-10, Clean-Code) and receive a consolidated pass/fail assessment.

## Key Features

- **Grade Threshold Assessment**: Enforces a minimum grade-level standard for your code.
- **Coverage Threshold Assessment**: Checks average coverage values against a specified percentage threshold.
- **Tool-Specific Filtering**: Include or exclude specific tool checks (e.g., `SOLID`, `OWASP-Top-10`, `Clean-Code`, etc.).
- **Flexible Configuration**: A `.codeleft` file and `config.json` can be customized to fine-tune how the tool runs and what it ignores.
- **History Tracking**: A `history.json` file logs prior results to track how grades evolve over time.

## Installation

1. **Download the Latest Release**  
   Download the appropriate binary from the [GitHub Releases](https://github.com/henrylamb/codeleft-cli/releases). Look for an archive that matches your operating system and architecture, for example: `codeleft-cli_Linux_x86_64.tar.gz`.

2. **Unpack and Move to Path**
   ```bash
   tar -xzf codeleft-cli_Linux_x86_64.tar.gz
   chmod +x codeleft-cli
   sudo mv codeleft-cli /usr/local/bin
   ```
   Adjust the file names and paths based on your OS and needs.

3. **Verify Installation**
   ```bash
   codeleft-cli --version
   ```
   You should see output indicating **version 1.0.3** (or whichever current version you installed).

## Quick Start

To run **codeleft-cli** in a repository that contains (or will contain) `.codeleft`, `config.json`, and `history.json`, simply call:

```bash
codeleft-cli
```

- **If `.codeleft`, `history.json`, or `config.json` do not exist in the current directory**, `codeleft-cli` will create them using default values.
- **By default**, the tool verifies if you have set any thresholds or chosen any tools. If no flags are passed, the CLI will simply generate the needed files and exit with a success message.

## Configuration Files

### `.codeleft`
A placeholder file that signals the tool to treat the current directory as the project root for **codeleft-cli** operations.
> _If `.codeleft` does not exist, the tool generates it automatically._

### `history.json`
Stores a log of prior assessments, enabling the CLI to track and filter the latest results.
> _If `history.json` does not exist, the tool generates it as an empty array: `[]`._

### `config.json`
Contains the configuration specifics, such as ignored files/folders or other advanced settings.
> _If `config.json` does not exist, the tool generates it with default data. You can then modify this file to tweak how **codeleft-cli** runs, which paths it ignores, etc._

**Example** (Empty or minimal):
```json
{
  "Ignore": {
    "Folders": [],
    "Files": []
  }
}
```

You can customize which folders and files to ignore by populating these arrays.

## CLI Flags and Options

**codeleft-cli** supports several flags to control its behavior:

| Flag                  | Description                                                                                         | Default |
|-----------------------|-----------------------------------------------------------------------------------------------------|---------|
| `-threshold-grade`    | A string (e.g., `"A"`, `"B"`, etc.) that sets the minimum acceptable grade. If the latest grades are lower, the CLI fails. | *None*  |
| `-threshold-percent`  | An integer percentage (e.g., `80`) used as the minimum acceptable coverage. If the average coverage is below this, the CLI fails. | *None*  |
| `-tools`              | A comma-separated list of tools (e.g., `"SOLID,OWASP-Top-10,PR-Readiness"`) to include in the assessment. | *None*  |
| `-asses-grade`        | A boolean (either `true` or `false`) that determines if the grade threshold should be assessed.                    | `false` |
| `-asses-coverage`     | A boolean (either `true` or `false`) that determines if the coverage threshold should be assessed.                 | `false` |
| `-version`            | If set, prints the current version of **codeleft-cli** and exits.                                                 | *None*  |

### Tooling Examples

The `-tools` flag can accept various tool names, including but not limited to:
- **SOLID**
- **OWASP-Top-10**
- **Clean-Code**
- **Functional-Coverage**
- **MISRA-C++**
- **PR-Ready**

Use them as follows:
```bash
codeleft-cli -tools "SOLID,OWASP-Top-10,PR Ready"
```

## Usage Examples

1. **Run with Grade Threshold**
   ```bash
   codeleft-cli -threshold-grade "A" -asses-grade=true
   ```
   This will create or update `.codeleft`, `config.json`, and `history.json` (if they don’t exist), and then check if the project’s latest grades meet or exceed `"A"`. If not, the CLI exits with a non-zero status.

2. **Run with Coverage Threshold**
   ```bash
   codeleft-cli -threshold-percent=80 -asses-coverage=true
   ```
   The CLI checks whether the project meets an 80% average coverage. If coverage is below 80%, the CLI fails.

3. **Combine Grade and Coverage Assessments**
   ```bash
   codeleft-cli -threshold-grade "B" -asses-grade=true -threshold-percent=70 -asses-coverage=true
   ```
   The CLI ensures both the grade is at least **B** and the coverage is at least **70%**. Fails if either check is not met.

4. **Specify Tools**
   ```bash
   codeleft-cli -tools "SOLID,OWASP-Top-10,Clean-Code" 
   ```
   Only these three tooling checks count toward the pass/fail logic, ignoring other categories in `history.json`.

5. **Check Version**
   ```bash
   codeleft-cli -version
   ```
   Displays `codeleft-cli Version 1.0.2` (or your installed version) and exits.

## Troubleshooting

1. **Missing `.codeleft` or `config.json`**
    - If these are missing, the CLI auto-creates them with default values. Adjust the newly created `config.json` if needed.
2. **Unexpected Failures**
    - Verify that you are passing the correct flags (`-asses-grade` or `-asses-coverage`) in conjunction with `-threshold-grade` or `-threshold-percent`.
3. **No Tools in Results**
    - Ensure your `-tools` flag matches the tool names in your `history.json`.
    - Example: If your `history.json` has results under `"OWASP-Top-10"`, specifying `-tools "OWASP-Top-10"` is required to include them.

## Contributing

We welcome community contributions! To contribute:

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/some-new-feature`).
3. Commit your changes (`git commit -am 'Add some feature'`).
4. Push to the branch (`git push origin feature/some-new-feature`).
5. Open a Pull Request.

Please ensure your changes pass any existing tests and adhere to our coding standards.

## License

This project is licensed under the [MIT License](LICENSE.md). Feel free to use, modify, and distribute this software in accordance with the license terms.

---

**Happy coding with codeleft-cli!** For further questions or discussion, feel free to open an issue or reach out via [GitHub Discussions](https://github.com/henrylamb/codeleft-cli/discussions).