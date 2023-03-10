package main

import (
	configuration "CFScanner/config"
	scan "CFScanner/scanner"
	utils "CFScanner/utils"
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var config = scan.ConfigStruct{
	Local_port:           0,
	Address_port:         "0",
	User_id:              "",
	Ws_header_host:       "",
	Ws_header_path:       "",
	Sni:                  "",
	Do_upload_test:       false,
	Min_dl_speed:         50.0,
	Min_ul_speed:         50.0,
	Max_dl_time:          -2.0,
	Max_ul_time:          -2.0,
	Max_dl_latency:       -1.0,
	Max_ul_latency:       -1.0,
	Fronting_timeout:     -1.0,
	Startprocess_timeout: -1.0,
	N_tries:              -1,
	No_vpn:               false,
}

// Program Info
var (
	version  = "0.5"
	build    = "Custom"
	codename = "CFScanner , CloudFlare Scanner."
)

func Version() string {
	return version
}

// VersionStatement returns a list of strings representing the full version info.
func VersionStatement() string {
	return strings.Join([]string{
		"CFScanner ", Version(), " (", codename, ") ", build, " (", runtime.Version(), " ", runtime.GOOS, "/", runtime.GOARCH, ")",
	}, "")
}

func main() {
	var threads int
	var configPath string
	var noVpn bool
	var subnetsPath string
	var doUploadTest bool
	var nTries int
	var minDLSpeed float64
	var minULSpeed float64
	var maxDLTime float64
	var maxULTime float64

	var startProcessTimeout int
	var frontingTimeout float64
	var maxDLLatency float64
	var maxULLatency float64

	var bigIPList []string

	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: codename,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(VersionStatement())
			if !noVpn {
				utils.CreateDir(scan.CONFIGDIR)
				// configfilepath := configPath
			}
			utils.CreateDir(scan.RESULTDIR)
			if err := configuration.CreateInterimResultsFile(scan.INTERIM_RESULTS_PATH, nTries); err != nil {
				fmt.Printf("Error creating interim results file: %v\n", err)
			}
			threadsCount := threads

			var IPLIST []string
			if subnetsPath != "" {
				subnetFilePath := subnetsPath
				subnetFile, err := os.Open(subnetFilePath)
				if err != nil {
					panic(err)
				}
				defer subnetFile.Close()

				scanner := bufio.NewScanner(subnetFile)
				for scanner.Scan() {
					IPLIST = append(IPLIST, strings.TrimSpace(scanner.Text()))
				}
				if err := scanner.Err(); err != nil {
					panic(err)
				}
				// } else {
				// 	cidrList = readCidrsFromAsnLookup()
				// }
			}
			testConfig := configuration.CreateTestConfig(configPath, time.Duration(startProcessTimeout), doUploadTest,
				minDLSpeed, minULSpeed, maxDLTime, maxULTime,
				frontingTimeout, maxDLLatency, maxULLatency,
				nTries, noVpn)

			var nTotalIPs int

			for _, ips := range IPLIST {
				numIPs := utils.GetNumIPsInCIDR(ips)
				nTotalIPs += numIPs
			}

			bigIPList = utils.IPParser(IPLIST)

			fmt.Println("Total Threads : ", utils.Colors.OKBLUE, threads, utils.Colors.ENDC)
			fmt.Printf("Starting to scan %d IPS.\n", nTotalIPs)
			fmt.Println("---------------------------")
			scan.Scanner(&testConfig, bigIPList, threadsCount)
			fmt.Println("Results Written in :", scan.INTERIM_RESULTS_PATH)
		},
	}
	rootCmd.PersistentFlags().IntVar(&threads, "threads", 1, "Number of threads to use for parallel computing")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "The path to the config file. For confg file example, see https://github.com/MortezaBashsiz/CFScanner/blob/main/bash/ClientConfig.json")
	rootCmd.PersistentFlags().BoolVar(&noVpn, "novpn", false, "If passed, test without creating vpn connections")
	rootCmd.PersistentFlags().StringVar(&subnetsPath, "subnets", "", "(optional) The path to the custom subnets file. each line should be in the form of ip.ip.ip.ip/subnet_mask or ip.ip.ip.ip. If not provided, the program will read the cidrs from asn lookup")
	rootCmd.PersistentFlags().BoolVar(&doUploadTest, "upload-test", false, "If True, upload test will be conducted")
	rootCmd.PersistentFlags().IntVar(&nTries, "tries", 1, "Number of times to try each IP. An IP is marked as OK if all tries are successful")
	rootCmd.PersistentFlags().Float64Var(&minDLSpeed, "download-speed", 50, "Minimum acceptable download speed in kilobytes per second")
	rootCmd.PersistentFlags().Float64Var(&minULSpeed, "upload-speed", 50, "Maximum acceptable upload speed in kilobytes per second")
	rootCmd.PersistentFlags().Float64Var(&maxDLTime, "download-time", 2, "Maximum (effective, excluding http time) time to spend for each download")
	rootCmd.PersistentFlags().Float64Var(&maxULTime, "upload-time", 2, "Maximum (effective, excluding http time) time to spend for each upload")
	rootCmd.PersistentFlags().Float64Var(&frontingTimeout, "fronting-timeout", 1.0, "Maximum time to wait for fronting response")
	rootCmd.PersistentFlags().Float64Var(&maxDLLatency, "download-latency", 2.0, "Maximum allowed latency for download")
	rootCmd.PersistentFlags().Float64Var(&maxULLatency, "upload-latency", 2.0, "Maximum allowed latency for download")
	rootCmd.PersistentFlags().IntVar(&startProcessTimeout, "startprocess-timeout", 5, "")

	err := rootCmd.Execute()
	cobra.CheckErr(err)

}
