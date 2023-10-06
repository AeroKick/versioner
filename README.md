# Versioner

### What is Versioner?
> Versioner is a dead simple CLI tool designed to easily bump version numbers using a symantic versioning system(Major.Minor.Patch).
It has a super basic config file (versioner.json) that lives in your project root directory.

### Installation

Versioner can be installed via Go using `go install github.com/BetterKICK/versioner@latest`

### Configuration

Configuring versioner is quite simple, here is what a basic `versioner.json` file may look like:
```json versioner.json
[
  {"file": "package.json", "field": "version"},
  {"file": "manifest.json", "field": "versionNumber"},
]
```

In the above versioner.json file, we are telling versioner that there are 2 JSON files in which it needs to update the version number when ran.
(Note that versioner currently only supports top level fields, and JSON files)

### Usage
Once you have a config file setup, you can simply run `versioner` in your CLI, this will start the CLI and ask which type of version bump(major, minor, patch) you would like to make.

Once you press enter, versioner will go through each file in the configuration, and bump the version number by the type you selected. Each file is bumped independently of eachother, so you could have one on 1.2.1 and one on 2.4.2, and they would be bumped to 1.2.2 and 2.4.3 respectively.



