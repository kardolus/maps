# Maps CLI

Maps CLI is an AI-powered command-line tool that simplifies searching Google Maps. Using the Google Places API, it
refines search queries and breaks down large results into manageable chunks. The AI automatically handles query
breakdown and filters out irrelevant results, making it easier than ever to search for locations. Results can be saved
to a JSON file or printed directly to the console.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
    - [Basic Usage](#basic-usage)
    - [Environment Variables](#environment-variables)
    - [Autocompletion](#autocompletion)
        - [Bash](#bash)
        - [Zsh](#zsh)
        - [Fish](#fish)
        - [PowerShell](#powershell)
- [Flags](#flags)
- [Example](#example)
- [AI-Powered Query Breakdown](#ai-powered-query-breakdown)
- [Configuration](#configuration)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **AI-Powered Query Refinement**: Automatically break down large Google Maps search queries into smaller, manageable
  sub-queries.
- **Smart Filtering**: Uses AI to ensure only relevant results are returned (e.g., filtering out non-relevant
  locations).
- **Google Places API Integration**: Fetches locations directly using the Google Places API based on a search query.
- **Save or Print Results**: Optionally save results to a JSON file or print them to the console.
- **Shell Autocompletion**: Autocompletion support for Bash, Zsh, Fish, and Powershell.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/kardolus/maps
   cd maps
   ```

2. Build the CLI:

   ```bash
   ./script/install.sh
   ```

## Usage

### Basic Usage

You can run the CLI with a search query and API key:

```bash
maps --query "Whole Foods in USA" --api-key YOUR_GOOGLE_PLACES_API_KEY
```

If you want to save the results to a JSON file:

```bash
maps --query "Whole Foods in USA" --api-key YOUR_GOOGLE_PLACES_API_KEY --output results.json
```

### Environment Variables

You can set the API key via an environment variable instead of passing it through the command line:

```bash
export GOOGLE_API_KEY=YOUR_GOOGLE_PLACES_API_KEY
maps --query "Whole Foods in USA"
```

### Autocompletion

To enable autocompletion for your shell, run the following command:

#### Bash

```bash
maps completion bash > /etc/bash_completion.d/maps
source /etc/bash_completion.d/maps
```

#### Zsh

```bash
maps completion zsh > "${fpath[1]}/_maps"
source ~/.zshrc
```

#### Fish

```bash
maps completion fish | source
```

#### PowerShell

```powershell
maps completion powershell | Out-String | Invoke-Expression
```

## Flags

- `--query, -q`: The search query (default: `"Whole Foods in USA"`).
- `--api-key`: Google Places API key. Can also be set via the `GOOGLE_API_KEY` environment variable.
- `--output, -o`: Optional file path to write the JSON response.

## Example

```bash
maps --query "Parks in San Francisco" --api-key YOUR_API_KEY --output parks_sf.json
```

## AI-Powered Query Breakdown

The Maps CLI integrates with an AI service to break down larger queries into sub-queries and apply filters. For example,
a search for "Whole Foods in USA" will break down the query into more manageable sub-regions like:

- "Whole Foods in New York"
- "Whole Foods in California"

The AI also filters results to ensure they match the specified terms (e.g., "Whole Foods Market").

## Configuration

You can place a configuration file named `config.yaml` in the `bin` directory. It can be used to store API keys or other
persistent settings. The configuration file supports the following fields:

```yaml
api-key: YOUR_GOOGLE_PLACES_API_KEY
query: "Whole Foods in USA"
output: "results.json"
```

## Testing

To run unit tests:

```bash
./scripts/unit.sh
```

## Contributing

Feel free to submit pull requests or file issues if you encounter any bugs or have suggestions.


