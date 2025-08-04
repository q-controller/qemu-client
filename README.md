# QEMU client

This project provides a lightweight Go package for interacting with the QEMU command-line tool. It simplifies launching and managing QEMU virtual machines (VMs) by supporting a curated subset of QEMU options, enabling:

1. **VM Initialization**: Start VM instances with customizable hardware parameters (e.g., memory, CPU, disk).
2. **Cloud-Init Support**: Handle cloud-init ISO files for automated VM configuration.
3. **Flexible Networking**: Configure networking for:
   - **VM-to-VM communication**
   - **VM-to-host communication**
   - **Internet access**
   
   Networking is implemented using the `vmnet` framework on macOS and TAP devices on Linux, ensuring platform-specific compatibility.

## Getting Started

### Prerequisites
- **QEMU**: Installed on your system (`qemu-system-x86_64` or equivalent).
- **Disk Images**: A QEMU-compatible disk image (e.g., `.qcow2`) for each VM.
- **Cloud Init**: If cloud init to be used, then some packages are required to be installed on a system:
  * genisoimage (Linux)
  * cdrtools (macOS)
