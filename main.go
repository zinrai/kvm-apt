package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DomainDisk struct {
	XMLName xml.Name `xml:"disk"`
	Source  struct {
		File string `xml:"file,attr"`
	} `xml:"source"`
}

type Domain struct {
	XMLName xml.Name     `xml:"domain"`
	Devices struct {
		Disks []DomainDisk `xml:"disk"`
	} `xml:"devices"`
}

func main() {
	var imagePath string
	var packages string

	flag.StringVar(&imagePath, "image", "", "Path to the KVM image")
	flag.StringVar(&packages, "packages", "", "Comma-separated list of packages to install")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 || packages == "" {
		fmt.Println("Usage: kvm-apt [-image <image_path>] -packages <package1,package2,...> <vm_name>")
		os.Exit(1)
	}

	vmName := args[0]

	if !isVMStopped(vmName) {
		fmt.Printf("The VM '%s' is currently running. Please stop it before customizing.\n", vmName)
		os.Exit(1)
	}

	var targetImage string
	var err error

	if imagePath != "" {
		targetImage = imagePath
		fmt.Printf("Using specified image: %s\n", targetImage)
		if err := verifyImageBelongsToVM(targetImage, vmName); err != nil {
			fmt.Printf("Verification failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		targetImage, err = getVMDiskPath(vmName)
		if err != nil {
			fmt.Printf("Failed to get disk path for VM '%s': %v\n", vmName, err)
			os.Exit(1)
		}
		fmt.Printf("Using first disk of VM '%s': %s\n", vmName, targetImage)
	}

	fmt.Printf("Packages to install: %s\n", packages)

	packageList := strings.Split(packages, ",")
	err = customizeKVMImage(targetImage, packageList)
	if err != nil {
		fmt.Printf("KVM image customization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("KVM image customization completed successfully.")
}

func customizeKVMImage(imagePath string, packages []string) error {
	args := []string{
		"virt-customize",
		"-a", imagePath,
		"--install", strings.Join(packages, ","),
	}

	fmt.Printf("Executing command: sudo %s\n", strings.Join(args, " "))

	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("sudo virt-customize command failed: %w", err)
	}

	return nil
}

func isVMStopped(vmName string) bool {
	cmd := exec.Command("sudo", "virsh", "list", "--name", "--state-running")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get list of running VMs: %v\n", err)
		return false
	}

	runningVMs := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, vm := range runningVMs {
		if vm == vmName {
			return false
		}
	}
	return true
}

func getVMDiskPath(vmName string) (string, error) {
	cmd := exec.Command("sudo", "virsh", "dumpxml", vmName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get VM XML: %w", err)
	}

	var domain Domain
	err = xml.Unmarshal(output, &domain)
	if err != nil {
		return "", fmt.Errorf("failed to parse VM XML: %w", err)
	}

	if len(domain.Devices.Disks) == 0 {
		return "", fmt.Errorf("no disks found for VM")
	}

	return domain.Devices.Disks[0].Source.File, nil
}

func verifyImageBelongsToVM(imagePath, vmName string) error {
	disks, err := getVMDisks(vmName)
	if err != nil {
		return fmt.Errorf("failed to get disks for VM '%s': %w", vmName, err)
	}

	for _, disk := range disks {
		if disk == imagePath {
			fmt.Printf("Image '%s' belongs to VM '%s'\n", imagePath, vmName)
			return nil
		}
	}

	return fmt.Errorf("the specified image '%s' is not connected to VM '%s'", imagePath, vmName)
}

func getVMDisks(vmName string) ([]string, error) {
	cmd := exec.Command("sudo", "virsh", "dumpxml", vmName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get VM XML: %w", err)
	}

	var domain Domain
	err = xml.Unmarshal(output, &domain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VM XML: %w", err)
	}

	var disks []string
	for _, disk := range domain.Devices.Disks {
		disks = append(disks, disk.Source.File)
	}

	return disks, nil
}
