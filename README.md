# kvm-apt

`kvm-apt` is a command-line tool designed to simplify the process of installing packages on KVM (Kernel-based Virtual Machine) images. It allows users to install packages on either a specified VM or a direct image file.

## Features

- Install packages on a KVM VM by specifying the VM name
- Install packages on a KVM image file directly
- Verify that the specified image belongs to a known VM
- Ensure the target VM is stopped before making changes

## Notes

- The tool requires sudo privileges to run virsh commands and virt-customize.
- When not using the `-image` option, the tool will automatically use the first disk of the specified VM.
- When using the `-image` option, the tool will verify that the image belongs to the specified VM before proceeding.
- The tool will not make changes to running VMs. Ensure the target VM is stopped before using this tool.

## Requirements

- Sudo access
- virsh installed on the system
- virt-customize tool

## Installation

Build the project:

```
$ go build
```

## Usage

### Installing packages on a VM:

```
$ ./kvm-apt -packages <package1,package2,...> <vm_name>
```

Example:
```
$ ./kvm-apt -packages python3,golang,vim ubuntu20-04
```

### Installing packages on a specific image:

```
$ ./kvm-apt -image <path_to_image> -packages <package1,package2,...> <vm_name>
```

Example:
```
$ ./kvm-apt -image /path/to/debian.qcow2 -packages python3-consul,nginx debian10
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://opensource.org/license/mit) for details.
