package main

import (
	"encoding/json"
	"fmt"
	"github.com/kardolus/maps/client"
	"github.com/kardolus/maps/http"
	"github.com/kardolus/maps/llm"
	"github.com/kardolus/maps/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "maps",
	Short: "Fetch locations using Google Places API",
	Long:  "Fetch locations using Google Places API and optionally write the JSON response to a file",
	RunE:  run,
}

var validShellArgs = []string{"bash", "zsh", "fish", "powershell"}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate autocompletion script",
	Long:  "Generate autocompletion script for your shell (bash, zsh, fish, or powershell)",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validShellArgs, cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			fmt.Println("Unsupported shell type")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("query", "q", "Whole Foods In USA", "Search query")
	viper.BindPFlag("query", rootCmd.PersistentFlags().Lookup("query"))

	rootCmd.PersistentFlags().String("api-key", "", "Google Places API key")
	viper.BindPFlag("api-key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindEnv("api-key", "GOOGLE_API_KEY")

	rootCmd.PersistentFlags().StringP("output", "o", "", "Output file to write the JSON response")
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("bin")
	viper.AutomaticEnv()

	rootCmd.AddCommand(completionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// TODO bootstrap this for testing
func run(cmd *cobra.Command, args []string) error {
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return fmt.Errorf("missing Google Places API key, set it via --api-key flag or GOOGLE_API_KEY environment variable")
	}

	c := client.New(http.New().WithRetries(3), apiKey).WithTimeout(5000)

	query := viper.GetString("query")
	fmt.Printf("Fetching locations for query: %s\n", query)

	gpt, err := llm.NewChatGPTClient()
	if err != nil {
		return err
	}

	ai := llm.New(gpt, &utils.Utils{})

	if err := ai.ClearHistory(); err != nil {
		return err
	}

	queries, err := ai.GenerateSubQueries(query)
	if err != nil {
		return err
	}

	if err := ai.ClearHistory(); err != nil {
		return err
	}

	contains, matches, err := ai.GenerateFilter(query)
	if err != nil {
		return err
	}

	locations, err := c.FetchAllLocations(queries, contains, matches)
	if err != nil {
		return err
	}

	outputFile := viper.GetString("output")
	if outputFile != "" {
		fmt.Printf("Writing results to file: %s\n", outputFile)

		data, err := json.MarshalIndent(locations, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal locations: %w", err)
		}

		// Write the JSON to the file
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	} else { // If no output file is specified, print the results to stdout
		data, err := json.MarshalIndent(locations, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal locations: %w", err)
		}
		fmt.Println(string(data))
	}

	return nil
}
