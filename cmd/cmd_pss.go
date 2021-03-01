package cmd

import (
	"fmt"
	"strconv"

	"github.com/cinus-ue/securekit/kit/pass"
	"github.com/cinus-ue/securekit/util"
	"github.com/urfave/cli/v2"
)

var Pss = &cli.Command{
	Name:   "pss",
	Usage:  "Generate secure, random password",
	Action: PassAction,
}

func PassAction(*cli.Context) error {
	val := util.GetInput("Password Length:")
	length, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	password := pass.GenerateRandomString(true, true, length)
	fmt.Printf("Your new password is: %s\n", password)
	return nil
}