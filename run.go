package main

import (
	"fmt"
	"io"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

// CalmOptions holds options for the run command
type CalmOptions struct {
	CommandToRun string
	Args         []string
	Memory       string
	CPU          string
	User         string
	Debug        bool
}

// NewRunCommand returns the run command with all of its options populated
func NewRunCommand(in io.Reader, out, errorOut io.Writer) *cobra.Command {
	calmOpts := CalmOptions{}

	calmCommand := &cobra.Command{
		Use:     "calm [flags] <command> <command arg1>...",
		Short:   "run a program using cgroup resource limits",
		Example: "calm google-chrome",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			calmOpts.CommandToRun = args[0]
			calmOpts.Args = args[1:]

			err := calmOpts.Run(in, out, errorOut)
			if err != nil {
				errorOut.Write([]byte(err.Error() + "\n"))
			}
		},
	}

	flags := calmCommand.PersistentFlags()
	flags.BoolVar(&calmOpts.Debug, "debug", false, debugUsage)
	flags.StringVarP(&calmOpts.User, "user", "u", "", userUsage)
	flags.StringVarP(&calmOpts.Memory, "memory", "m", "", memoryUsage)
	flags.StringVarP(&calmOpts.CPU, "cpu", "c", "", cpuUsage)

	return calmCommand
}

// Run actually runs the process that needs to be calm'd down
func (r *CalmOptions) Run(in io.Reader, out, errorOut io.Writer) error {

	config, err := createConfig()
	if err != nil {
		return fmt.Errorf("could not create config: %v", err)
	}

	if r.Memory != "" {
		config.Set(MEMORY_CONFIG_KEY, r.Memory)
	}

	if r.CPU != "" {
		config.Set(CPU_CONFIG_KEY, r.CPU)
	}

	if r.User != "" {
		config.Set(USER_KEY, r.User)
	}

	logIfDebug(errorOut, r.Debug,
		"Memory limits: %s\tCPU limits: %s\tUser: %s\n",
		config.GetString(MEMORY_CONFIG_KEY),
		config.GetString(CPU_CONFIG_KEY),
		config.GetString(USER_KEY),
	)

	spec, err := createCgroupSpecFromConfig(config)
	if err != nil {
		return fmt.Errorf("could not create cgroup limit specification: %v", err)
	}

	err = enterCgroup("calm", spec)
	if err != nil {
		return fmt.Errorf("could not enter cgroup: %v", err)
	}

	cmd := exec.Command(r.CommandToRun, r.Args...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errorOut

	username := config.GetString(USER_KEY)
	usr, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("could not look up configured user: %s: %v", username, err)
	}

	uid, err := strconv.ParseUint(usr.Uid, 10, 32)
	if err != nil {
		return err
	}
	gid, err := strconv.ParseUint(usr.Gid, 10, 32)
	if err != nil {
		return err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not start command: %v", err)
	}

	err = cmd.Wait()
	return err
}

func logIfDebug(out io.Writer, debug bool, format string, a ...interface{}) {
	if debug {
		yellowFormatted := color.YellowString(format, a...)
		fmt.Fprint(out, yellowFormatted)
	}
}
