# Redfish development tools - Virtual Redfish BMC

## Redfish Simulation Emulator

After reviewing a few Redfish simulation tools, our choice, moving forward,
was a tool called [sushy-tools](https://docs.openstack.org/sushy-tools/latest/)
developed by the Openstack community. This tool simulates the Redfish protocols
and provides the development community with independent access and testing of
the Redfish protocol implementations. This tool is actively being enhanced and
provides support for uefi boot. As such, one may encounter temporary hiccups
with the code if one tries to use the latest code, thus we provide the git
commit sha1 for the code we tested in the prerequisites sections that follow.

## About Sushy-Tools

The sushy-tools tool set includes two emulators - static and dynamic. We have
chosen to use the dynamic emulator as we want to use the libvirt backend to
mimic baremetal nodes behind sushy-emulator (Redfish BMC). The sushy-emulator
command-line tool contains functionality that is similar to the Virtual BMC
tool except it uses the Redfish frontend protocols rather than IPMI. Refer
to the diagram in the sushy-tools.png file which accompanies this
documentation for an illustration of the toolset.

The sushy-emulator provides many Redfish resources that help the developers
kickstart their efforts. These include:

- Systems Resource
- Managers Resource
- Indicators Resource
- Virtual Media Resource

## **Host Node Installation**

The host node needs to have the operating system installed along with some
additional tools. This section provides instructions for preparing Ubuntu
hosted on Bare Metal as well as describing some of the prerequisites and
tested software and hardware components.

### **Installation Prerequisites**

Before we begin the installation there are a few prerequisites that should be
considered, such as:

- Are you installing the sushy-tools directly on an existing node or host or
hosting the tool inside a virtual machine or instance?
- While it is possible to host the tools on Windows, our assumption and favored
choice was to install on a Linux with qemu/libvirtd support.
- Hardware is a matter of choice as longs as the hardware supports virtual
systems.

### **Tested Hardware and Software**

For our development purposes and based on what was available to us at the time,
we selected and tested the following hardware and software.

#### **Tested Hardware**

Dell R640 PowerEdge servers with the following accessories:

- Dual Intel Xeon 6126 2.6GHz CPUs
- 192GB - 2666Mhz RAM
- 800GB RAID10 storage
- 10GB bonded Intel NIC

However, we believe the minimum/functional requirements are much more relaxed.

An x86 hardware server/laptop with:

- around 6 vCPUs available
- 8GB RAM, for creating the emulator VM as well as the target VM
- 200GB storage, for creating root disks for the VMs
- optional to have a NIC, since VM-VM communication will be based on SW bridges

#### **Tested Base Operating System**

- Ubuntu 18.04
- Default server installation
- Other APT packages installed included: ``net-tools``, ``zip``, ``unzip``,
``git``, ``qemu``, and ``libvirtd`` with dependencies

## **Ubuntu 18.04 Hosted on Bare Metal**

### **OS installation**

Installation of Ubuntu 18.04 be done either using an ISO image mounted via
virtual media or physical media, adjust according to your host hardware.

Once you have the operating system installed, the following instructions should
be followed to install the requisite packages for the Redfish emulator,  the
Redfish emulator itself and the backend virtual node.

#### **Installation of requisite APT packages**

```bash
sudo apt update -y
sudo apt install -y git python3-setuptools qemu-kvm libvirt-bin virtinst \
python3-flask python3-requests nginx libvirt-daemon virt-manager libvirt \
libvirt-python libvirt-client python3-libvirt python-libvirt
```

This completes the process for Ubuntu 18.04 hosted on bare metal, please
proceed to the **Configuration of the Redfish Emulator and vbmc-node** section.

## Ubuntu 18.04 Hosted on a VM

As indicated above in the installation prerequisites, one has the option of
installing sushy-tools directly on the host system or within a virtual machine
hosted by the host system. Initially, we went with the latter, simply because
it was easier to tear down and build up experimental Redfish simulation engine
without polluting our host node.
The ``sushy-tools`` is also available as a container image in the metal3
project. You can use [this image](https://quay.io/repository/metal3-io/sushy-tools)
to instantiate the setup using either Docker or Podman as your runtime.

### Building the Redfish VM on a host system

Note: You can use the [apache-wsgi-sushy-emulator](https://opendev.org/airship/airshipctl/src/branch/master/roles/apache-wsgi-sushy-emulator)
ansible role in the airshipctl repo to setup the emulator with/without
authentication, as a WSGI application behind an apache virtual host. This is
helpful specifically in environments where you want better scaling of your
testing infrastructure and is what Airship uses for testing.

For experimental/smaller setups the following method can also be used.

#### Get the files to create the VM

##### Download the Ubuntu 18.04 Server image

[ubuntu-18.04.1-server-amd64.iso](http://old-releases.ubuntu.com/releases/18.04.3/ubuntu-18.04.1-server-amd64.iso)

##### Download the Redfish_tools.zip archive and build the redfish emulator VM

[Redfish_tools.zip](https://github.com/dell-esg/tool-pkgs/blob/master/Redfish_tools.zip)
Note: tools-pkgs repository have been deprecated from dell-esg so this link leads to 404

##### Scp the redfish_tools zip to the host machine

Note: The host machine must be capable of running a qemu libvirt VM.

##### Extract the files from the zip archive

Unzip the files in your home directory and cd into that newly created
``Redfish_tools`` subdirectory.

##### Modify the redfish.cfg file

Make the appropriate changes for your domain / network

```bash
    user root
    password r00tm3
    timezone UTC
    hostname redfish.oss.labs
    gateway 192.168.122.1
    nameserver 8.8.8.8
    ntpserver 0.centos.pool.ntp.org
```

##### Adjust Redfish Admin VM Public IP

Change the IP and netmask below to the IP address and netmask for the Redfish
Admin VM on the Public API network

```bash
    ens3 192.168.122.10 255.255.255.0 1500
```

##### Run the deployment script to deploy the VM

This will be done using the cfg file and the path to Ubuntu image that you
downloaded earlier.

```bash
./deploy-redfish-vm.py redfish.cfg \
/var/lib/libvirt/images/ubuntu-18.04.1-server-amd64.iso
```

Optional -- You can watch the VM deployment using virt-viewer (if you
previously installed the virt-viewer APT package on your host system and have
X windows installed)

```bash
    virt-viewer redfish
```

##### When the VM has finished installing, start the VM

```bash
    virsh start redfish
```

##### SSH into the VM using the IP address assigned in step 5

```bash
    ssh root@192.168.122.10
```

### **Configuration of the Redfish Emulator and virtual node**

NOTE:If you deployed the Redfish infrastructure VM using the
"deploy-redfish-vm.py" script, you can skip to **Verify the sushy emulator
is working** as the script does the work in the next 3 sections for you.

#### **Configure and Install the Sushy-emulator**

```bash
git clone (https://opendev.org/openstack/sushy-tools.git)
cd sushy-tools/
python3 setup.py build
python3 setup.py install
```

#### **Update the redfishd and emulator.conf files**

```bash
scp localsystem//redfishd.service root@redfish_vm_ip://tmp
scp [localsystem://emulator.conf] root@redfish_vm_ip://tmp
vi /tmp/redfishd.service #adjust file for the redfish_vm_ip
vi /tmp/emulator.conf #adjust file for the redfish_vm_ip
mkdir -p /etc/redfish
cp /tmp/emulator.conf /etc/redfish/
cp /tmp/redfishd.service /etc/systemd/system
systemctl start redfishd
systemctl status redfishd
systemctl enable redfishd
```

#### **Build the virtual node**

```bash
tmpfile=$(mktemp /tmp/sushy-domain.XXXXXX)
virt-install --name virtual-node --ram 1024 --boot uefi --disk size=1 --vcpus 2\
--os-type linux --os-variant fedora28 --graphics vnc --print-xml > $tmpfile
virsh define --file $tmpfile
rm $tmpfile
```

#### **Verify the sushy emulator is working and the virtual-node was added**

```bash
curl -L 'http://192.168.122.10:8000/redfish/v1/Systems'
curl -L 'http://192.168.122.10:8000/redfish/v1/Systems/8e5b2dc4-0c1d-4509-af2f-7a4a8f2121a8'
```

Note: For virtual media boot, instead of using the IP address 192.168.122.10
(used above), ``localhost`` was used in the commands that follow.

#### Download the bionicpup64-8.0-uefi.iso image

[bionicpup64-8.0](http://distro.ibiblio.org/puppylinux/puppy-bionic/bionicpup64/bionicpup64-8.0-uefi.iso)

Note: Use the 64-bit image with 64-bit VMs as the 32-bit image will hang during
kernel initialization.

#### Upload the image to /tmp of the host node using your preferred scp tool

```bash
cp /tmp/bionicpup32-8.0-uefi.iso /var/www/nginxsite.com/public_html/mini.iso
```

One might also rename the bionicpup64-8.0-uefi.iso to mini.iso to match the
documentation

#### Use a browser to verify the image is downloadable from the webserver

With a browser goto the URL: ``http://localhost/mini.iso``
The browser should proceed to download the file

#### Build a UEFI bootable virtual node

```bash
virsh list --all
tmpfile=$(mktemp /tmp/sushy-domain.XXXXXX)
virt-install --name virtual-node --ram 1024 --boot uefi --disk size=1000 \
vcpus 2 --os-type linux --graphics - - vnc --print-xml > $tmpfile
virsh define --file $tmpfile
curl http://localhost:8000/redfish/v1/Systems/
```

#### Retrieve the system odata.id (47a3b9a3-3967-4d23-98d8-18de1c28e94f)

It is required in the commands below

```bash
curl -d '{"Image":"http://localhost/mini.iso", "Inserted": true}'\
-H "Content-Type: application/json" -X POST \
http://localhost:8000/redfish/v1/Managers/58893887-894-2487-2389-841168418919/VirtualMedia/Cd/Actions/VirtualMedia.InsertMedia-+
```

#### Mount the mini.iso

```bash
curl http://localhost:8000/redfish/v1/Managers/58893887-894-2487-2389-841168418919/VirtualMedia/Cd
```

#### Verify the image is mounted

Expect the following values in the returned data of the API call

- "Image": "mini.iso",
- "ConnectedVia": "URI", "Inserted": true

```bash
curl -X PATCH -H 'Content-Type: application/json' -d '{ "Boot":{\
"BootSourceOverrideTarget": "Cd", "BootSourceOverrideMode": "Uefi",\
"BootSourceOverrideEnabled": "Continuous" } }'\
http://localhost:8000/redfish/v1/Systems/47a3b9a3-3967-4d23-98d8-18de1c28e94f
```

#### This sets the BootSourceOverrideTarget,BootSourceOverrideMode

BootSourceOverrideEnabled fields for the vbmc-node**

```bash
curl http://localhost:8000/redfish/v1/Systems/47a3b9a3-3967-4d23-98d8-18de1c28e94f
```

#### To verify the BootSourceOverride fields are set correctly

```bash
curl -d '{"ResetType":"On"}' -H "Content-Type: application/json" -X POST\
http://localhost:8000/redfish/v1/Systems/47a3b9a3-3967-4d23-98d8-18de1c28e94f/Actions/ComputerSystem.Reset
```

#### Boot the node

Watch the system boot in the virt-manager console or via ``virt-viewer
virtual-node`` command

### **Some helpful links**

- [Openstack Sushy Storyboard](https://storyboard.openstack.org/#!/story/list)
- [Redfish development tools](https://docs.openstack.org/sushy-tools/latest/)
