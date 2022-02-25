package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"mdibl_cloud_control/datamodels"
	"mdibl_cloud_control/utils"
	"os"
	"strconv"
	"strings"

	"github.com/vaughan0/go-ini"
)

func main() {
	// Boolean flags
	var help,
		listInstances,
		listInstanceTypes,
		stopAllInstances,
		stopInstances,
		startAllInstances,
		startInstances,
		launchInstances *bool

	// String flags
	var awsConf *string

	// Declare boolean flags
	help = flag.Bool("help", false, "Show full help message")
	listInstances = flag.Bool("list-instances", false, "List all running instances")
	listInstanceTypes = flag.Bool("list-instance-types", false, "List available instance types for region")
	stopAllInstances = flag.Bool("stop-all-instances", false, "Stop all running instances")
	stopInstances = flag.Bool("stop-instances", false, "Stop instances specified in instance report")
	startAllInstances = flag.Bool("start-all-instances", false, "Start all stopped instances")
	startInstances = flag.Bool("start-instances", false, "Start all instances specified in instance report")
	launchInstances = flag.Bool("launch-instances", false, "Launch instances from a config file")

	// Declare string flags
	awsConf = flag.String("aws-config", ".aws/config", "Path to aws config folder.")
	flag.Parse()

	/* -------------------------------------------------------------------------
	 * Check for the help param and display help message if provided.
	 * ---------------------------------------------------------------------- */
	if *help {
		fmt.Println("\nWill show a help message eventually.\n")
		os.Exit(0)
	}

	/* -------------------------------------------------------------------------
	 * Proceed with other operations
	 * ---------------------------------------------------------------------- */

	// Get the default AWS region
	region, err := utils.DefaultAWSRegion(*awsConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Create new EC@ client with specified credentials
	creds, err := utils.CreateNewEC2ClientCredentials(*awsConf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ec2Client := utils.CreateNewEC2Client(creds, region)

	// List all instance types for the specified region
	if *listInstanceTypes {
		dryRun := false
		params := utils.CreateInstanceTypeOfferingFilterParams(dryRun)
		describeInstanceTypeOutput, _ := ec2Client.DescribeInstanceTypeOfferings(params)
		offerings := utils.ParseInstanceTypeOfferings(describeInstanceTypeOutput)
		outputFileName, err := utils.WriteInstanceTypeOfferings(region, offerings)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("\nOutput written to %s\n", outputFileName)
		os.Exit(0)
	}

	// List all instances in the user's account. Running, stopped or pending will be returned.
	if *listInstances {
		// Create filter parameters
		params := utils.CreateEC2InstanceFilterParams("instance-state-name", []string{"stopped", "running", "pending"})

		// Query AWS for all instances that match the filter params
		describeInstanceOutput, _ := ec2Client.DescribeInstances(params)

		// Parse instance details
		instanceReport := utils.ParseDescribeInstanceOutput(describeInstanceOutput)

		// Print instance details to screen
		utils.PrintEC2InstanceReport(instanceReport)

		// Write instance report to file
		outputFileName, err := utils.WriteInstanceDetailsReport(instanceReport, "all")
		fmt.Printf("\nOutput written to %s\n", outputFileName)
		if err != nil {
			panic(err)
		}
	}

	/* -------------------------------------------------------------------------
	 * Stop all running instances
	 * ---------------------------------------------------------------------- */
	if *stopAllInstances {
		// Get report of all running instances.
		// Create filter parameters. Only care about running instances
		params := utils.CreateEC2InstanceFilterParams("instance-state-name", []string{"running"})
		// Query AWS for all instances that match the filter params
		describeInstanceOutput, _ := ec2Client.DescribeInstances(params)

		// Print instance information
		report := utils.ParseDescribeInstanceOutput(describeInstanceOutput)
		if len(report.Instances) == 0 {
			// Abort if there are no instances to stop
			fmt.Println("No running instances")
			os.Exit(0)
		}

		// List running instances for user.
		// Build array of image ids at the same time.
		var instanceIDs = make([]string, 0)
		var response string

		// Prompt the user before proceeding
		fmt.Printf("\nThe following instances will be stopped:\n")
		fmt.Printf("----------------------------------------\n\n")
		for _, instance := range report.Instances {
			// Append id string.
			instanceIDs = append(instanceIDs, instance.InstanceID)
			// Display it to the user.
			fmt.Printf("Name: %s, ID: %s, Instance type: %s\n", instance.Name, instance.InstanceID, instance.InstanceType)
		}
		fmt.Printf("\nContinue: (y/n) ")
		fmt.Scanln(&response)

		// Stop all running instances
		if strings.ToLower(response) == "y" {
			fmt.Println("Stopping instances...")
			// Create stop instance param object
			stopInstanceParams := utils.CreateEC2StopInstanceParams(instanceIDs, true, false, false)
			// Throw away the response for now.
			_, err := ec2Client.StopInstances(stopInstanceParams)
			// Only panic on an actual error. Not dry run operation report.
			if err != nil && !strings.Contains(err.Error(), "DryRunOperation") {
				panic(err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Println("Bye!")
		}
		os.Exit(0)
	}

	/* -------------------------------------------------------------------------
	 * Stop all specified instances
	 * ---------------------------------------------------------------------- */
	if *stopInstances {
		if len(flag.Args()) == 0 {
			fmt.Println("Instance details file required but not supplied")
			os.Exit(1)
		}

		// Grab the instance config file from the command line.
		instanceReport := flag.Args()[0]

		// Make sure the config file exists. If configuration is not present, abort.
		if _, err := os.Stat(instanceReport); os.IsNotExist(err) {
			fmt.Printf("No instance report file found at: %s\n", instanceReport)
			os.Exit(1)
		}

		// Read json file
		jsonData, err := ioutil.ReadFile(instanceReport)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Unmarshal the json
		ec2ReportObj := datamodels.EC2InstanceReport{}
		err = json.Unmarshal(jsonData, &ec2ReportObj)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var response string
		// Warn user about stopping all images. Prompt for continue.
		// Generate the instance id array at the same time.
		var instanceIDs = make([]string, 0)
		fmt.Printf("\nThe following instances will be stopped:\n")
		fmt.Printf("----------------------------------------\n\n")
		for _, instance := range ec2ReportObj.Instances {
			// Append id string.
			instanceIDs = append(instanceIDs, instance.InstanceID)
			// Display it to the user.
			fmt.Printf("Name: %s, ID: %s, Instance type: %s\n", instance.Name, instance.InstanceID, instance.InstanceType)
		}
		fmt.Printf("\nContinue: (y/n) ")
		fmt.Scanln(&response)

		// Stop all running instances
		if strings.ToLower(response) == "y" {
			fmt.Println("Stopping specified instances...")
			// Create stop instance param object
			stopInstanceParams := utils.CreateEC2StopInstanceParams(instanceIDs, false, false, false)
			// Throw away the response for now.
			_, err := ec2Client.StopInstances(stopInstanceParams)
			// Only panic on an actual error. Not dry run operation report.
			if err != nil && !strings.Contains(err.Error(), "DryRunOperation") {
				panic(err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Println("Bye!")
		}
		os.Exit(0)
	}

	/* -------------------------------------------------------------------------
	 * Start all stopped instances.
	 * NOTE: This will not create new instances. Only start existing instances.
	 * ---------------------------------------------------------------------- */
	if *startAllInstances {
		// Create filter parameters. Only care about stopped instances
		params := utils.CreateEC2InstanceFilterParams("instance-state-name", []string{"stopped"})
		// Query AWS for all instances that match the filter params
		describeInstanceOutput, _ := ec2Client.DescribeInstances(params)

		// Print instance information
		report := utils.ParseDescribeInstanceOutput(describeInstanceOutput)
		if len(report.Instances) == 0 {
			// Abort if there are no instances to start
			fmt.Println("No stopped instances")
			os.Exit(0)
		}

		// List running instances for user.
		// Build array of image ids at the same time.
		var instanceIDs = make([]string, 0)
		var response string

		fmt.Printf("\nThe following instances will be started:\n")
		fmt.Printf("------------------------------------------\n\n")
		for _, instance := range report.Instances {
			// Append id string.
			instanceIDs = append(instanceIDs, instance.InstanceID)
			// Display it to the user.
			fmt.Printf("Name: %s, ID: %s, Instance type: %s\n", instance.Name, instance.InstanceID, instance.InstanceType)
		}
		fmt.Printf("\nContinue: (y/n) ")
		fmt.Scanln(&response)

		// Start all stopped instances
		if strings.ToLower(response) == "y" {
			fmt.Println("Starting all instances...")
			// Create stop instance param object
			startInstanceParams := utils.CreateEC2StartInstanceParams(instanceIDs, true)
			// Throw away the response for now.
			_, err := ec2Client.StartInstances(startInstanceParams)
			// Only panic on an actual error. Not dry run operation report.
			if err != nil && !strings.Contains(err.Error(), "DryRunOperation") {
				panic(err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Println("Bye!")
		}
		os.Exit(0)
	}

	/* -------------------------------------------------------------------------
	 * Start all stopped instances.
	 * NOTE: This will not create new instances. Only start existing instances.
	 * ---------------------------------------------------------------------- */
	if *startInstances {
		if len(flag.Args()) == 0 {
			fmt.Println("Instance details file required but not supplied")
			os.Exit(1)
		}

		// Grab the instance config file from the command line.
		instanceReport := flag.Args()[0]

		// Make sure the config file exists. If configuration is not present, abort.
		if _, err := os.Stat(instanceReport); os.IsNotExist(err) {
			fmt.Printf("No config file found at: %s\n", instanceReport)
			os.Exit(1)
		}

		// Read json file
		jsonData, err := ioutil.ReadFile(instanceReport)
		if err != nil {
			panic(err)
		}

		// Unmarshal the json
		ec2ReportObj := datamodels.EC2InstanceReport{}
		err = json.Unmarshal(jsonData, &ec2ReportObj)
		if err != nil {
			panic(err)
		}

		var response string
		// Warn user about stopping all images. Prompt for continue.
		// Generate the instance id array at the same time.
		var instanceIDs = make([]string, 0)
		fmt.Printf("\nThe following instances will be started:\n")
		fmt.Printf("----------------------------------------\n\n")
		for _, instance := range ec2ReportObj.Instances {
			// Append id string.
			instanceIDs = append(instanceIDs, instance.InstanceID)
			// Display it to the user.
			fmt.Printf("Name: %s, ID: %s, Instance type: %s\n", instance.Name, instance.InstanceID, instance.InstanceType)
		}
		fmt.Printf("\nContinue: (y/n) ")
		fmt.Scanln(&response)

		// Start all instances
		if strings.ToLower(response) == "y" {
			fmt.Println("Starting specified instances...")
			// Create stop instance param object
			startInstanceParams := utils.CreateEC2StartInstanceParams(instanceIDs, false)
			// Throw away the response for now.
			_, err := ec2Client.StartInstances(startInstanceParams)
			// Only panic on an actual error. Not dry run operation report.
			if err != nil && !strings.Contains(err.Error(), "DryRunOperation") {
				panic(err)
			}
			fmt.Println("Done!")
		} else {
			fmt.Println("Bye!")
		}
		os.Exit(0)
	}

	if *launchInstances {
		if len(flag.Args()) == 0 {
			fmt.Println("Instance config file required but not supplied")
			os.Exit(1)
		}

		// Grab the instance config file from the command line.
		instanceConf := flag.Args()[0]

		// Make sure the config file exists. If configuration is not present, abort.
		if _, err := os.Stat(instanceConf); os.IsNotExist(err) {
			fmt.Printf("No instance config file found at: %s\n", instanceConf)
			os.Exit(1)
		}

		// Load the config file and extract the parameters.
		configFile, err := ini.LoadFile(instanceConf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		amiID, _ := configFile.Get("instance", "ami_id")
		amiName, _ := configFile.Get("instance", "ami_name")
		instanceType, _ := configFile.Get("instance", "instance_type")
		countString, _ := configFile.Get("instance", "count")
		count, _ := strconv.ParseInt(countString, 10, 64)

		// Display launch request to user
		var response string
		fmt.Println("\nLaunch request details:")
		fmt.Printf("-----------------------\n\n")
		fmt.Printf("AMI Name: %s\nInstance type: %s\nCount: %d\n\n", amiName, instanceType, count)
		fmt.Printf("Continue: (y/n) ")
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" {
			// Create specified instances.
			createInstanceParams := utils.CreateEC2RunInstanceParams(amiID, instanceType, count)
			runResponse, err := ec2Client.RunInstances(createInstanceParams)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Generate a launch report and write it to disk.
			reportType := "launch"
			report := utils.GetInstanceDetails(runResponse)
			_, err = utils.WriteInstanceDetailsReport(report, reportType)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("Done!")
			os.Exit(0)
		}
		fmt.Println("Bye!")
		os.Exit(0)
	}
}
