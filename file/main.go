package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Madhu-1/gluster-csi-drivers/file/driver"
	"github.com/spf13/cobra"
)

var (
	endpoint   string
	nodeID     string
	glusterURL string
	username   string
	secret     string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "gluster-csi-driver",
		Short: "CSI based Gluster driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "node id")
	cmd.MarkPersistentFlagRequired("nodeid")

	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "CSI endpoint")

	cmd.PersistentFlags().StringVar(&glusterURL, "glusterurl", "", "glusterd2 endpoint")
	cmd.MarkPersistentFlagRequired("glusterurl")

	cmd.PersistentFlags().StringVar(&username, "username", "glustercli", "glusterd2 user name")
	//	cmd.MarkPersistentFlagRequired("username")

	cmd.PersistentFlags().StringVar(&secret, "secret", "", "glusterd2 user secret")
	//	cmd.MarkPersistentFlagRequired("secret")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {
	if endpoint == "" {
		endpoint = os.Getenv("CSI_ENDPOINT")
	}

	d := driver.NewDriver(nodeID, endpoint, glusterURL, username, secret)
	d.Run()
}
