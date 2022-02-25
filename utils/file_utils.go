package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mdibl_cloud_control/datamodels"
	"os"
	"strings"
	"time"
)

func WriteInstanceTypeOfferings(region string, instanceTypes []string) (string, error) {
	var err error

	// Format the output file name.
	filename := fmt.Sprintf("instance_types_%s.txt", region)
	outfile, err := os.Create(filename)
	if err != nil {
		return filename, err
	}
	writer := bufio.NewWriter(outfile)
	for _, elem := range instanceTypes {
		writer.WriteString(fmt.Sprintf("%s\n", elem))
	}
	writer.Flush()
	return filename, err
}

func WriteInstanceDetailsReport(report datamodels.EC2InstanceReport, reportType string) (string, error) {
	timeString := strings.Replace(time.Now().Format("2006-01-02 15:04:05"), " ", "_", -1)
	// Format the report file name.
	reportName := fmt.Sprintf("%s_instance_details_%s.json", reportType, timeString)
	outputJSON, err := json.Marshal(report)
	if err != nil {
		return reportName, err
	}
	return reportName, ioutil.WriteFile(reportName, outputJSON, os.ModePerm)
}
