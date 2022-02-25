# MDIBL Cloud Control

A command line tool for interacting with AWS EC2 API.

## Getting started
	git clone git@github.com:cwilson28/mdibl_cloud_control.git mdibl_cloud_control
	cd mdibl_cloud_control

Add your AWS credentials and region to .aws/config for authentication with the AWS API.

	aws_access_key_id=YOUR_ACCESS_KEY
	aws_secret_access_key=YOUR_SECRET_ACCESS_KEY
	region=YOUR_INSTANCE_REGION (e.g., us-east-2)
	output=json

## Usage
To run the tool:

	go run mdibl_cloud_control [option] [args]

Available options are:

	--help	Show full help message
	--list-instance-types	List all available instance types for your region. Output is written to the local file instance_types_<region>.txt.
	--list-instances	List all EC2 instances. Will include all instances with states "running", "stopped" and "pending".
	--stop-all-instances	Stops all running instances
	--stop-instances <path_to_instance_report>	Stop all instances specified in instance report.
	--start-all-instances	Start all stopped instances.
	--start-instances <path_to_instance_report>	Start all instances specified in instance report.
	--launch-instances <path_to_instance_config> Launch instances from a config file.

A launch instance config file has the format

	[instance]
	ami_id=ID_OF_TARGET_AMI
	ami_name=NAME_OF_AMI
	instance_type=(e.g., t2.micro)
	region=YOUR_REGION
	count=NUMBER_OF_MACHINES_TO_LAUNCH

An empty config file is provided as instance.config. 

After executing the launch command, the instance details will be written to a local instance report.
