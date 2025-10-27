package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/treerootboy/anyalert/pkg/database"
	"github.com/treerootboy/anyalert/pkg/userstore"
)

const (
	defaultDBPath = "anyalert.db"
)

var (
	userDBPath     string
	userName       string
	userSlack      string
	userYoudu      string
	userPhone      string
	userSMS        string
	userJSONOutput bool
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Manage user metadata for different notification channels`,
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user",
	Long:  `Add a new user with notification channel information`,
	Example: `  anyalert user add --name user1 --slack user1@example.com --youdu 10232 --phone +8613800138000`,
	Run:     runUserAdd,
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  `List all users in the database`,
	Example: `  anyalert user list
  anyalert user list --json`,
	Run: runUserList,
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a user by name",
	Long:  `Get user details by username`,
	Example: `  anyalert user get --name user1
  anyalert user get --name user1 --json`,
	Run: runUserGet,
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing user",
	Long:  `Update user notification channel information`,
	Example: `  anyalert user update --name user1 --slack updated@example.com
  anyalert user update --name user1 --slack new@example.com --phone +8613800138001`,
	Run: runUserUpdate,
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  `Delete a user from the database`,
	Example: `  anyalert user delete --name user1`,
	Run:     runUserDelete,
}

func init() {
	// user command flags
	userAddCmd.Flags().StringVar(&userDBPath, "db", defaultDBPath, "Path to SQLite database")
	userAddCmd.Flags().StringVar(&userName, "name", "", "User name (required)")
	userAddCmd.Flags().StringVar(&userSlack, "slack", "", "Slack email or username")
	userAddCmd.Flags().StringVar(&userYoudu, "youdu", "", "Youdu account ID")
	userAddCmd.Flags().StringVar(&userPhone, "phone", "", "Phone number")
	userAddCmd.Flags().StringVar(&userSMS, "sms", "", "SMS phone number")
	userAddCmd.MarkFlagRequired("name")

	userListCmd.Flags().StringVar(&userDBPath, "db", defaultDBPath, "Path to SQLite database")
	userListCmd.Flags().BoolVar(&userJSONOutput, "json", false, "Output in JSON format")

	userGetCmd.Flags().StringVar(&userDBPath, "db", defaultDBPath, "Path to SQLite database")
	userGetCmd.Flags().StringVar(&userName, "name", "", "User name (required)")
	userGetCmd.Flags().BoolVar(&userJSONOutput, "json", false, "Output in JSON format")
	userGetCmd.MarkFlagRequired("name")

	userUpdateCmd.Flags().StringVar(&userDBPath, "db", defaultDBPath, "Path to SQLite database")
	userUpdateCmd.Flags().StringVar(&userName, "name", "", "User name (required)")
	userUpdateCmd.Flags().StringVar(&userSlack, "slack", "", "Slack email or username")
	userUpdateCmd.Flags().StringVar(&userYoudu, "youdu", "", "Youdu account ID")
	userUpdateCmd.Flags().StringVar(&userPhone, "phone", "", "Phone number")
	userUpdateCmd.Flags().StringVar(&userSMS, "sms", "", "SMS phone number")
	userUpdateCmd.MarkFlagRequired("name")

	userDeleteCmd.Flags().StringVar(&userDBPath, "db", defaultDBPath, "Path to SQLite database")
	userDeleteCmd.Flags().StringVar(&userName, "name", "", "User name (required)")
	userDeleteCmd.MarkFlagRequired("name")

	userCmd.AddCommand(userAddCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userGetCmd)
	userCmd.AddCommand(userUpdateCmd)
	userCmd.AddCommand(userDeleteCmd)
	rootCmd.AddCommand(userCmd)
}

func runUserAdd(cmd *cobra.Command, args []string) {
	// 使用共享数据库配置
	dbConfig := database.Config{
		Path: userDBPath,
	}
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := userstore.NewSQLiteStoreWithDB(db.GetConnection())
	if err != nil {
		fmt.Printf("Failed to create user store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user := &userstore.User{
		Name:  userName,
		Slack: userSlack,
		Youdu: userYoudu,
		Phone: userPhone,
		SMS:   userSMS,
	}

	if err := store.Create(user); err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User created successfully (ID: %d)\n", user.ID)
	printUser(user)
}

func runUserList(cmd *cobra.Command, args []string) {
	// 使用共享数据库配置
	dbConfig := database.Config{
		Path: userDBPath,
	}
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := userstore.NewSQLiteStoreWithDB(db.GetConnection())
	if err != nil {
		fmt.Printf("Failed to create user store: %v\n", err)
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

	if userJSONOutput {
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

func runUserGet(cmd *cobra.Command, args []string) {
	// 使用共享数据库配置
	dbConfig := database.Config{
		Path: userDBPath,
	}
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := userstore.NewSQLiteStoreWithDB(db.GetConnection())
	if err != nil {
		fmt.Printf("Failed to create user store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(userName)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	if userJSONOutput {
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

func runUserUpdate(cmd *cobra.Command, args []string) {
	// 使用共享数据库配置
	dbConfig := database.Config{
		Path: userDBPath,
	}
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := userstore.NewSQLiteStoreWithDB(db.GetConnection())
	if err != nil {
		fmt.Printf("Failed to create user store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(userName)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	// Only update fields that were explicitly provided
	if cmd.Flags().Changed("slack") {
		user.Slack = userSlack
	}
	if cmd.Flags().Changed("youdu") {
		user.Youdu = userYoudu
	}
	if cmd.Flags().Changed("phone") {
		user.Phone = userPhone
	}
	if cmd.Flags().Changed("sms") {
		user.SMS = userSMS
	}

	if err := store.Update(user); err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("User updated successfully")
	printUser(user)
}

func runUserDelete(cmd *cobra.Command, args []string) {
	// 使用共享数据库配置
	dbConfig := database.Config{
		Path: userDBPath,
	}
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := userstore.NewSQLiteStoreWithDB(db.GetConnection())
	if err != nil {
		fmt.Printf("Failed to create user store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	user, err := store.GetByName(userName)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		os.Exit(1)
	}

	if err := store.Delete(user.ID); err != nil {
		fmt.Printf("Failed to delete user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User '%s' deleted successfully\n", userName)
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
