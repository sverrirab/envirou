package cmd

import (
	"errors"
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/crypt"
	"github.com/sverrirab/envirou/pkg/output"
)

var encryptStdout bool

const minimumPassphraseLength = 12

var encryptCmd = &cobra.Command{
	Use:   "encrypt [VALUE]",
	Short: "Encrypt a value for use in profiles or .env files",
	Long: `Encrypt a value with a passphrase and print an enc:v1:... token that can
be pasted into a [profile:x] section or a .env file. The value is prompted
for when not given as an argument (recommended: arguments end up in shell
history).

The first use creates the key material file (crypt.ini) in the config
directory. Back that file up: encrypted values cannot be recovered
without it.`,
	GroupID: "configuration",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := setupOrGetKey()
		value := ""
		if len(args) == 1 {
			value = args[0]
		} else {
			var err error
			value, err = crypt.ReadPassphrase("Value to encrypt: ")
			exitOnError(err)
		}
		token, err := crypt.Encrypt(key, value)
		exitOnError(err)
		if encryptStdout {
			fmt.Println(token)
		} else {
			output.Printf("Paste this into your profile or .env file:\n%s\n", token)
		}
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt TOKEN",
	Short: "Decrypt an enc:v1: token (displays the secret in clear text)",
	Long: `Verify an encrypted value by decrypting it and printing the plaintext.
Note that the plaintext is displayed in clear text on the terminal.`,
	GroupID: "configuration",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !crypt.IsEncrypted(args[0]) {
			output.Printf("Not an encrypted value (expected %s... prefix)\n", crypt.TokenPrefix)
			os.Exit(1)
		}
		key, err := ensureKey()
		exitOnError(err)
		plaintext, err := crypt.Decrypt(key, args[0])
		if err != nil {
			if errors.Is(err, crypt.ErrMalformedToken) {
				output.Printf("Malformed token\n")
			} else {
				output.Printf("Cannot decrypt: value was encrypted with different key material (crypt.ini mismatch?)\n")
			}
			os.Exit(1)
		}
		output.Printf("%s\n", plaintext)
	},
}

// setupOrGetKey returns a verified key, creating the key material on first
// use (with a double passphrase prompt).
func setupOrGetKey() []byte {
	dir := config.GetDefaultConfigFileFolder()
	_, err := crypt.LoadMaterial(dir)
	if err == nil {
		key, err := ensureKey()
		exitOnError(err)
		return key
	}
	if !errors.Is(err, crypt.ErrNotInitialized) {
		exitOnError(err)
	}

	output.Printf("Setting up encryption (first use).\n")
	passphrase, err := crypt.ReadPassphrase("New passphrase: ")
	exitOnError(err)
	exitOnError(validateNewPassphrase(passphrase))
	repeat, err := crypt.ReadPassphrase("Repeat passphrase: ")
	exitOnError(err)
	if passphrase != repeat {
		exitOnError(errors.New("passphrases do not match"))
	}
	_, key, err := crypt.CreateMaterial(dir, passphrase, crypt.DefaultIterations)
	exitOnError(err)
	output.Printf("Created %s - back this file up! Encrypted values cannot be recovered without it.\n", crypt.MaterialPath(dir))
	app.cryptKey = key
	return key
}

func validateNewPassphrase(passphrase string) error {
	if utf8.RuneCountInString(passphrase) < minimumPassphraseLength {
		return fmt.Errorf("passphrase must be at least %d characters", minimumPassphraseLength)
	}
	return nil
}

func exitOnError(err error) {
	if err != nil {
		output.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	encryptCmd.Flags().BoolVar(&encryptStdout, "stdout", false, "Print the bare token to stdout (for scripting; use with envirou, not ev)")
	addCommand(encryptCmd)
	addCommand(decryptCmd)
}
