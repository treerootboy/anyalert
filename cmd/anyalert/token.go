package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/treerootboy/anyalert/pkg/database"
	"github.com/treerootboy/anyalert/pkg/token"
)

var (
	tokenDBPath     string
	tokenDesc       string
	tokenExpiresIn  string
	tokenID         string
	tokenJSONOutput bool
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
	Long:  `Manage API tokens for authentication`,
}

var tokenGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new token",
	Long:  `Generate a new API token with optional expiration`,
	Example: `  anyalert token generate --description "API token for service A"
  anyalert token generate --description "Temporary token" --expires-in 24h
  anyalert token generate --description "Test token" --json`,
	Run: runTokenGenerate,
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tokens",
	Long:  `List all tokens in the database`,
	Example: `  anyalert token list
  anyalert token list --json`,
	Run: runTokenList,
}

var tokenRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a token",
	Long:  `Revoke a token by its ID`,
	Example: `  anyalert token revoke --id abc123`,
	Run:     runTokenRevoke,
}

func init() {
	// token generate command flags
	tokenGenerateCmd.Flags().StringVar(&tokenDBPath, "db", defaultDBPath, "Path to SQLite database")
	tokenGenerateCmd.Flags().StringVar(&tokenDesc, "description", "", "Token description (required)")
	tokenGenerateCmd.Flags().StringVar(&tokenExpiresIn, "expires-in", "", "Expiration duration (e.g., 24h, 7d, 30d)")
	tokenGenerateCmd.Flags().BoolVar(&tokenJSONOutput, "json", false, "Output in JSON format")
	tokenGenerateCmd.MarkFlagRequired("description")

	// token list command flags
	tokenListCmd.Flags().StringVar(&tokenDBPath, "db", defaultDBPath, "Path to SQLite database")
	tokenListCmd.Flags().BoolVar(&tokenJSONOutput, "json", false, "Output in JSON format")

	// token revoke command flags
	tokenRevokeCmd.Flags().StringVar(&tokenDBPath, "db", defaultDBPath, "Path to SQLite database")
	tokenRevokeCmd.Flags().StringVar(&tokenID, "id", "", "Token ID to revoke (required)")
	tokenRevokeCmd.MarkFlagRequired("id")

	tokenCmd.AddCommand(tokenGenerateCmd)
	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenRevokeCmd)
	rootCmd.AddCommand(tokenCmd)
}

func openTokenDatabase(dbPath string) (*token.Manager, func(), error) {
	dbConfig := database.Config{
		Path: dbPath,
	}

	db, err := database.New(dbConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	mgr := token.NewManager(db.GetConnection())

	cleanup := func() {
		db.Close()
	}

	return mgr, cleanup, nil
}

func runTokenGenerate(cmd *cobra.Command, args []string) {
	mgr, cleanup, err := openTokenDatabase(tokenDBPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 解析过期时间
	var expiresInDuration *time.Duration
	if tokenExpiresIn != "" {
		duration, err := time.ParseDuration(tokenExpiresIn)
		if err != nil {
			fmt.Printf("Invalid expiration duration: %v\n", err)
			fmt.Println("Examples: 24h, 168h (7 days), 720h (30 days)")
			os.Exit(1)
		}
		expiresInDuration = &duration
	}

	// 生成 token
	tok, err := mgr.Generate(tokenDesc, expiresInDuration)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		os.Exit(1)
	}

	// 输出 token
	if tokenJSONOutput {
		data, _ := json.MarshalIndent(tok, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("\n✅ Token generated successfully!")
		fmt.Println("\n📋 Token Information:")
		fmt.Printf("  ID:          %s\n", tok.ID)
		fmt.Printf("  Value:       %s\n", tok.Value)
		fmt.Printf("  Description: %s\n", tok.Description)
		fmt.Printf("  Created At:  %s\n", tok.CreatedAt.Format(time.RFC3339))
		if tok.ExpiresAt != nil {
			fmt.Printf("  Expires At:  %s\n", tok.ExpiresAt.Format(time.RFC3339))
		} else {
			fmt.Printf("  Expires At:  Never expires\n")
		}

		fmt.Println("\n💡 Tips:")
		fmt.Println("  Token has been saved to the database.")
		fmt.Println("  Make sure 'token.enabled' is set to true in the config to enable authentication.")
	}
}

func runTokenList(cmd *cobra.Command, args []string) {
	mgr, cleanup, err := openTokenDatabase(tokenDBPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	tokens := mgr.List()

	if len(tokens) == 0 {
		fmt.Println("📭 No tokens found")
		fmt.Println("\n💡 Tip: Use 'anyalert token generate' to create a new token")
		return
	}

	if tokenJSONOutput {
		data, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("\n📋 Token List (%d total):\n\n", len(tokens))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tDescription\tCreated At\tExpires At\tStatus")
		fmt.Fprintln(w, "---\t---\t---\t---\t---")

		for _, tok := range tokens {
			expiresAt := "Never expires"
			status := "✅ Valid"

			if tok.ExpiresAt != nil {
				expiresAt = tok.ExpiresAt.Format("2006-01-02 15:04:05")
				if time.Now().After(*tok.ExpiresAt) {
					status = "❌ Expired"
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				tok.ID,
				tok.Description,
				tok.CreatedAt.Format("2006-01-02 15:04:05"),
				expiresAt,
				status,
			)
		}

		w.Flush()
		fmt.Println()
	}
}

func runTokenRevoke(cmd *cobra.Command, args []string) {
	mgr, cleanup, err := openTokenDatabase(tokenDBPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 撤销 token
	if err := mgr.RevokeByID(tokenID); err != nil {
		fmt.Printf("Failed to revoke token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Token %s has been revoked\n", tokenID)
}
