package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/treerootboy/anyalert/pkg/database"
	"github.com/treerootboy/anyalert/pkg/token"
)

const (
	defaultDBPath = "anyalert.db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generateToken(os.Args[2:])
	case "list":
		listTokens(os.Args[2:])
	case "revoke":
		revokeToken(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AnyAlert Token Management Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  tokenmgr <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  generate  Generate a new token")
	fmt.Println("  list      List all tokens")
	fmt.Println("  revoke    Revoke a token")
	fmt.Println("\nExamples:")
	fmt.Println("  tokenmgr generate --description \"API token for service A\"")
	fmt.Println("  tokenmgr generate --description \"Temporary token\" --expires-in 24h")
	fmt.Println("  tokenmgr list")
	fmt.Println("  tokenmgr list --json")
	fmt.Println("  tokenmgr revoke --id abc123")
}

func openDatabase(dbPath string) (*token.Manager, func(), error) {
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

func generateToken(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	description := fs.String("description", "", "Token description (required)")
	expiresIn := fs.String("expires-in", "", "Expiration duration (e.g., 24h, 7d, 30d)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)

	if *description == "" {
		fmt.Println("Error: --description is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	mgr, cleanup, err := openDatabase(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 解析过期时间
	var expiresInDuration *time.Duration
	if *expiresIn != "" {
		duration, err := time.ParseDuration(*expiresIn)
		if err != nil {
			fmt.Printf("Invalid expiration duration: %v\n", err)
			fmt.Println("Examples: 24h, 168h (7 days), 720h (30 days)")
			os.Exit(1)
		}
		expiresInDuration = &duration
	}

	// 生成 token
	tok, err := mgr.Generate(*description, expiresInDuration)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		os.Exit(1)
	}

	// 输出 token
	if *jsonOutput {
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

func listTokens(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)

	mgr, cleanup, err := openDatabase(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	tokens := mgr.List()

	if len(tokens) == 0 {
		fmt.Println("📭 No tokens found")
		fmt.Println("\n💡 Tip: Use 'tokenmgr generate' to create a new token")
		return
	}

	if *jsonOutput {
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

func revokeToken(args []string) {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	tokenID := fs.String("id", "", "Token ID to revoke (required)")

	fs.Parse(args)

	if *tokenID == "" {
		fmt.Println("Error: --id is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	mgr, cleanup, err := openDatabase(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 撤销 token
	if err := mgr.RevokeByID(*tokenID); err != nil {
		fmt.Printf("Failed to revoke token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Token %s has been revoked\n", *tokenID)
}
