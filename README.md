The expectation is that this command line tool will take in a directory of files for which in the root there is a file called .codeleft. If this file doesn't exist along with a history.json file or a config.json file then these will be created with defualt data. The config defualt is below:

```json

```

The history.json file created will be an empty array. 

The flags that can be used for this command line tool are as follows:

- -threshold-grade: This flag sets the point at which the latest grades will have match so that the general assessment fails or not.
- -threshold-percent: This flag sets the numerical average that the latest grades must be at to pass the test.
- -assessment: This flag takes in the values true or false and will determine if the assessment is run or not. This is not recommended to run as it could lead to long build times and that the assessment should take in place in the IDE. The default and recommended flag is false
- -tools: This flag takes in a list of values as strings which determines which types of tooling will or will not be included in the pass or fail metric. Below are some of the options:
- "SOLID"
- "OWASP-Top-10"
- "Clean-Code"
- "Functional-Coverage"
- "MISRA-C++"
- "PR-Readiness"
- 