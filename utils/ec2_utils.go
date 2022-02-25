package utils

import (
	"fmt"
	"mdibl_cloud_control/datamodels"
	"net/url"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/vaughan0/go-ini"
)

/* ---
 * Get default AWS region from config file
 * --- */
func DefaultAWSRegion(awsConfigFile string) (string, error) {
	var err error
	var region string

	// Make sure the config file exists. If configuration is not present, abort.
	if _, err := os.Stat(awsConfigFile); os.IsNotExist(err) {
		err = fmt.Errorf("No config file found at: %s\n", awsConfigFile)
		return region, err
	}

	// Load the config file and extract the parameters.
	paramFile, err := ini.LoadFile(awsConfigFile)
	if err != nil {
		return region, err
	}
	region, _ = paramFile.Get("default", "region")
	return region, nil
}

/* ---
 * Create a new AWS EC2 client credentials
 * --- */
func CreateNewEC2ClientCredentials(awsConfigFile string) (*credentials.Credentials, error) {
	var err error
	var creds *credentials.Credentials

	// Make sure the config file exists. If configuration is not present, abort.
	if _, err := os.Stat(awsConfigFile); os.IsNotExist(err) {
		err = fmt.Errorf("No config file found at: %s\n", awsConfigFile)
		return creds, err
	}

	// Load the config file and extract the parameters.
	paramFile, err := ini.LoadFile(awsConfigFile)
	if err != nil {
		return creds, err
	}

	keyID, _ := paramFile.Get("default", "aws_access_key_id")
	secretKey, _ := paramFile.Get("default", "aws_secret_access_key")
	return credentials.NewStaticCredentials(keyID, secretKey, ""), nil
}

/* ---
 * Create a new AWS EC2 client
 * --- */
func CreateNewEC2Client(creds *credentials.Credentials, region string) *ec2.EC2 {
	// Create a vanilla session
	mySession := session.Must(session.NewSession())
	// Create a EC2 client with additional configuration supplied by user.
	return ec2.New(mySession, aws.NewConfig().WithCredentials(creds).WithRegion(region))
}

/* -----------------------------------------------------------------------------
 * Functions for creating various AWS filter/request parameter objects.
 * -------------------------------------------------------------------------- */

/* ---
 * Create EC2 instance type offering filter params
 * --- */
func CreateInstanceTypeOfferingFilterParams(dryRun bool) *ec2.DescribeInstanceTypeOfferingsInput {
	// An array of aws.Strings (e.g., pointers to strings)
	// filterValues := make([]*string, 0)
	// filterValues = append(filterValues, aws.String(region))

	// Return aws filter parameter object
	return &ec2.DescribeInstanceTypeOfferingsInput{
		DryRun: aws.Bool(dryRun),
		// Filters: []*ec2.Filter{
		// 	&ec2.Filter{
		// 		Name:   aws.String("location"),
		// 		Values: filterValues,
		// 	},
		// },
	}
}

func ParseInstanceTypeOfferings(output *ec2.DescribeInstanceTypeOfferingsOutput) []string {
	var offerings = make([]string, 0)

	for _, elem := range output.InstanceTypeOfferings {
		offerings = append(offerings, *elem.InstanceType)
	}
	sort.Strings(offerings)
	return offerings
}

/* ---
 * Create EC2 instance filter params
 * --- */
func CreateEC2InstanceFilterParams(filterName string, filters []string) *ec2.DescribeInstancesInput {
	// An array of aws.Strings (e.g., pointers to strings)
	filterValues := make([]*string, 0)

	// Convert all user supplied filters to aws.String
	for _, f := range filters {
		filterValues = append(filterValues, aws.String(f))
	}

	// Return aws filter parameter object
	return &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String(filterName),
				Values: filterValues,
			},
		},
	}
}

func CreateEC2InstanceIDsParams(report datamodels.EC2InstanceReport) *ec2.DescribeInstancesInput {
	// An array of aws.Strings (e.g., pointers to strings)
	instanceIDs := make([]*string, 0)

	// Convert all user supplied filters to aws.String
	for _, instance := range report.Instances {
		instanceIDs = append(instanceIDs, aws.String(instance.InstanceID))
	}

	// Return aws filter parameter object
	return &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}
}

/* ---
 * Create EC2 start instance params
 * --- */
func CreateEC2StartInstanceParams(ids []string, dryrun bool) *ec2.StartInstancesInput {
	// An array of aws.Strings (e.g., pointers to strings)
	instanceIDs := make([]*string, 0)

	// Convert all user supplied filters to aws.String
	for _, id := range ids {
		instanceIDs = append(instanceIDs, aws.String(id))
	}

	// Return aws filter parameter object
	return &ec2.StartInstancesInput{
		InstanceIds: instanceIDs,
		DryRun:      aws.Bool(dryrun),
	}
}

/* ---
 * Create EC2 stop instance params
 * --- */
func CreateEC2StopInstanceParams(ids []string, dryrun, force, hibernate bool) *ec2.StopInstancesInput {
	// An array of aws.Strings (e.g., pointers to strings)
	instanceIDs := make([]*string, 0)

	// Convert all user supplied filters to aws.String
	for _, id := range ids {
		instanceIDs = append(instanceIDs, aws.String(id))
	}

	// Return aws filter parameter object
	return &ec2.StopInstancesInput{
		InstanceIds: instanceIDs,
		DryRun:      aws.Bool(dryrun),
		Force:       aws.Bool(force),
		Hibernate:   aws.Bool(hibernate),
	}
}

/* ---
 * Create EC2 run instance params
 * --- */
func CreateEC2RunInstanceParams(amiID, instanceType string, count int64) *ec2.RunInstancesInput {
	// Create run instance input for the specified AMI ID and instance type
	return &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: aws.String(instanceType),
		MinCount:     aws.Int64(count),
		MaxCount:     aws.Int64(count),
	}
}

// -----------------------------------------------------------------------------
// Functions for working with results of AWS API results
// -----------------------------------------------------------------------------

/* ---
 * Create a list of EC2Instance details.
 * --- */
func ParseDescribeInstanceOutput(results *ec2.DescribeInstancesOutput) datamodels.EC2InstanceReport {
	report := datamodels.EC2InstanceReport{}
	instances := make([]datamodels.EC2InstanceDetails, 0)

	// Results is a JSON(ish) object with the following form:
	// {reservations[{instances:[...]}]}
	// Loop over all reservation indexes
	for idx, _ := range results.Reservations {
		// Loop over all instances for each index
		for _, instance := range results.Reservations[idx].Instances {

			// We need to see if the Name is one of the tags. It's not always
			// present and not required in Ec2.
			name := "None"
			for _, keys := range instance.Tags {
				if *keys.Key == "Name" {
					name = url.QueryEscape(*keys.Value)
				}
			}

			// For now, just get the instance name, instance id, the instance type,
			// the public and private IPs if present
			details := datamodels.EC2InstanceDetails{}
			if &name != nil {
				details.Name = name
			}
			if instance.InstanceId != nil {
				details.InstanceID = *instance.InstanceId
			}
			if instance.InstanceType != nil {
				details.InstanceType = *instance.InstanceType
			}
			if instance.State.Name != nil {
				details.InstanceState = *instance.State.Name
			}
			if instance.PrivateIpAddress != nil {
				details.PrivateIP = *instance.PrivateIpAddress
			}
			if instance.PublicIpAddress != nil {
				details.PublicIP = *instance.PublicIpAddress
			}

			instances = append(instances, details)
		}
	}
	report.Instances = instances
	return report
}

/* ---
 * Print EC2Instance details to console.
 * --- */
func PrintEC2InstanceReport(report datamodels.EC2InstanceReport) {
	printString := "Name: %s\nInstanceID: %s\nInstanceType: %s\nInstance State: %s\nPublicIP %s\nPrivateIP: %s\n\n"
	fmt.Println("\nInstances")
	fmt.Println("---------")
	for _, elem := range report.Instances {
		fmt.Printf(printString, elem.Name, elem.InstanceID, elem.InstanceType, elem.InstanceState, elem.PublicIP, elem.PrivateIP)
	}
}

/* ---
 * Get the instance details from a launch.
 * --- */
func GetInstanceDetails(reservation *ec2.Reservation) datamodels.EC2InstanceReport {
	report := datamodels.EC2InstanceReport{}
	instances := make([]datamodels.EC2InstanceDetails, 0)

	// Results is a JSON(ish) object with the following form:
	// {reservations[{instances:[...]}]}
	// Loop over all reservation indexes
	for _, instance := range reservation.Instances {

		// We need to see if the Name is one of the tags. It's not always
		// present and not required in Ec2.
		name := "None"
		for _, keys := range instance.Tags {
			if *keys.Key == "Name" {
				name = url.QueryEscape(*keys.Value)
			}
		}

		// For now, just get the instance name, instance id, the instance type,
		// the public and private IPs if present
		details := datamodels.EC2InstanceDetails{}
		if &name != nil {
			details.Name = name
		}
		if instance.InstanceId != nil {
			details.InstanceID = *instance.InstanceId
		}
		if instance.InstanceType != nil {
			details.InstanceType = *instance.InstanceType
		}
		if instance.State.Name != nil {
			details.InstanceState = *instance.State.Name
		}
		if instance.PrivateIpAddress != nil {
			details.PrivateIP = *instance.PrivateIpAddress
		}
		if instance.PublicIpAddress != nil {
			details.PublicIP = *instance.PublicIpAddress
		}

		instances = append(instances, details)
	}
	report.Instances = instances
	return report
}
