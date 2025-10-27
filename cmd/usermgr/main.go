package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/treerootboy/anyalert/pkg/userstore"
)

const (
	defaultDBPath = "users.db"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add":
		addUser(os.Args[2:])
	case "list":
		listUsers(os.Args[2:])
	case "get":
		getUser(os.Args[2:])
	case "update":
		updateUser(os.Args[2:])
	case "delete":
		deleteUser(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AnyAlert User Management Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  usermgr <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  add     Add a new user")
	fmt.Println("  list    List all users")
	fmt.Println("  get     Get a user by name")
	fmt.Println("  update  Update an existing user")
	fmt.Println("  delete  Delete a user")
	fmt.Println("\nExamples:")
	fmt.Println("  usermgr add -name user1 -slack user1@example.com -youdu 10232 -phone +8613800138000")
	fmt.Println("  usermgr list")
	fmt.Println("  usermgr get -name user1")
	fmt.Println("  usermgr update -name user1 -slack updated@example.com")
	fmt.Println("  usermgr delete -name user1")
}

func addUser(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	name := fs.String("name", "", "User name (required)")
	slack := fs.String("slack", "", "Slack email or username")
	youdu := fs.String("youdu", "", "Youdu account ID")
	phone := fs.String("phone", "", "Phone number")
	sms := fs.String("sms", "", "SMS phone number")

	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	store, err := userstore.NewSQLiteStore(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user := &userstore.User{
		Name:  *name,
		Slack: *slack,
		Youdu: *youdu,
		Phone: *phone,
		SMS:   *sms,
	}

	if err := store.Create(user); err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User created successfully (ID: %d)\n", user.ID)
	printUser(user)
}

func listUsers(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)

	store, err := userstore.NewSQLiteStore(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	users, err := store.List()
	if err != nil {
		fmt.Printf("Failed to list users: %v\n", err)
		os.Exit(1)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return
	}

	if *jsonOutput {
		data, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tSlack\tYoudu\tPhone\tSMS")
	fmt.Fprintln(w, "--\t----\t-----\t-----\t-----\t---")
	for _, u := range users {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			u.ID, u.Name, u.Slack, u.Youdu, u.Phone, u.SMS)
	}
	w.Flush()
}

func getUser(args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	name := fs.String("name", "", "User name (required)")
	jsonOutput := fs.Bool("json", false, "Output in JSON format")

	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	store, err := userstore.NewSQLiteStore(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(*name)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		data, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	printUser(user)
}

func updateUser(args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	name := fs.String("name", "", "User name (required)")
	slack := fs.String("slack", "", "Slack email or username")
	youdu := fs.String("youdu", "", "Youdu account ID")
	phone := fs.String("phone", "", "Phone number")
	sms := fs.String("sms", "", "SMS phone number")

	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	store, err := userstore.NewSQLiteStore(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(*name)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	// Track which flags were explicitly set
	flagsSet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

	// Only update fields that were explicitly provided
	if flagsSet["slack"] {
		user.Slack = *slack
	}
	if flagsSet["youdu"] {
		user.Youdu = *youdu
	}
	if flagsSet["phone"] {
		user.Phone = *phone
	}
	if flagsSet["sms"] {
		user.SMS = *sms
	}

	if err := store.Update(user); err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("User updated successfully")
	printUser(user)
}

func deleteUser(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	dbPath := fs.String("db", defaultDBPath, "Path to SQLite database")
	name := fs.String("name", "", "User name (required)")

	fs.Parse(args)

	if *name == "" {
		fmt.Println("Error: -name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	store, err := userstore.NewSQLiteStore(*dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(*name)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	if err := store.Delete(user.ID); err != nil {
		fmt.Printf("Failed to delete user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User '%s' deleted successfully\n", *name)
}

func printUser(user *userstore.User) {
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("ID:         %d\n", user.ID)
	fmt.Printf("Name:       %s\n", user.Name)
	if user.Slack != "" {
		fmt.Printf("Slack:      %s\n", user.Slack)
	}
	if user.Youdu != "" {
		fmt.Printf("Youdu:      %s\n", user.Youdu)
	}
	if user.Phone != "" {
		fmt.Printf("Phone:      %s\n", user.Phone)
	}
	if user.SMS != "" {
		fmt.Printf("SMS:        %s\n", user.SMS)
	}
	fmt.Printf("Created:    %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:    %s\n", user.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("─────────────────────────────────────")
}
