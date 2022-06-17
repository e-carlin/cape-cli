package cmd

import (
	"fmt"
	"net/url"

	"github.com/capeprivacy/cli/capetest"
	czip "github.com/capeprivacy/cli/zip"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [directory | zip file] [input]",
	Short: "test your function with Cape",

	RunE: Test,
}

func init() {
	rootCmd.AddCommand(testCmd)
}

func wsURL(origURL string) string {
	u, _ := url.Parse(origURL)
	u.Scheme = "ws"

	return u.String()
}

func Test(cmd *cobra.Command, args []string) error {
	u, err := cmd.Flags().GetString("url")
	if err != nil {
		return fmt.Errorf("flag not found: %w", err)
	}

	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return fmt.Errorf("flag not found: %w", err)
	}

	if len(args) != 2 {
		if err := cmd.Usage(); err != nil {
			return err
		}

		return nil
	}

	fnZip, err := czip.Create(args[0])
	if err != nil {
		return err
	}

	input := []byte(args[1])
	res, err := test(capetest.TestRequest{Function: fnZip, Input: input}, wsURL(u), insecure)
	if err != nil {
		return err
	}

	if _, err := cmd.OutOrStdout().Write(res.Message); err != nil {
		return err
	}

	return nil
}

var test = capetest.CapeTest
