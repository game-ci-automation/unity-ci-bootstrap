import { DefaultAzureCredential } from '@azure/identity';
import { ComputeManagementClient } from '@azure/arm-compute';
import { NetworkManagementClient } from '@azure/arm-network';
import { ResourceManagementClient } from '@azure/arm-resources';

/**
 * Ensure Resource Group exists, create if not
 * @param {ResourceManagementClient} resourceClient - Azure Resource Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region (e.g., 'eastus')
 * @returns {Promise<void>} - No return value, throws on error
 */
async function ensureResourceGroup(resourceClient, resourceGroupName, location) {
    console.log(`Checking resource group: ${resourceGroupName}`);

    try {
        await resourceClient.resourceGroups.get(resourceGroupName);
        console.log(`Resource group ${resourceGroupName} already exists`);
    } catch (error) {
        if (error.statusCode === 404) {
            console.log(`Creating resource group ${resourceGroupName} in ${location}...`);
            await resourceClient.resourceGroups.createOrUpdate(resourceGroupName, {
                location: location
            });
            console.log(`Resource group ${resourceGroupName} created`);
        } else {
            throw error;
        }
    }
}

/**
 * Create Public IP resource
 * @param {NetworkManagementClient} networkClient - Azure Network Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region
 * @param {string} vmName - VM name (used for resource naming)
 * @returns {Promise<object>} Public IP object
 */
async function createPublicIP(networkClient, resourceGroupName, location, vmName) {
    const publicIPName = `${vmName}-ip`;

    try {
        console.log('Creating public IP...');
        const publicIPParams = {
            location: location,
            publicIPAllocationMethod: 'Static',
            sku: {
                name: 'Basic'
            }
        };

        const publicIP = await networkClient.publicIPAddresses.beginCreateOrUpdateAndWait(
            resourceGroupName,
            publicIPName,
            publicIPParams
        );

        console.log(`Public IP created: ${publicIP.ipAddress}`);

        return publicIP;
    } catch (error) {
        console.error(`Public IP creation failed for ${publicIPName}:`, error.message);
        console.error('Error details:', {
            code: error.code,
            statusCode: error.statusCode
        });
        throw new Error(`Failed to create Public IP: ${error.message}`);
    }
}

/**
 * Create Virtual Network and Subnet
 * @param {NetworkManagementClient} networkClient - Azure Network Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region
 * @param {string} vmName - VM name (used for resource naming)
 * @returns {Promise<object>} Virtual Network object
 */
async function createVirtualNetwork(networkClient, resourceGroupName, location, vmName) {
    const vnetName = `${vmName}-vnet`;
    const subnetName = `${vmName}-subnet`;

    console.log('Creating virtual network...');
    const vnetParams = {
        location: location,
        addressSpace: {
            addressPrefixes: ['10.0.0.0/16']
        },
        subnets: [{
            name: subnetName,
            addressPrefix: '10.0.0.0/24'
        }]
    };

    const vnet = await networkClient.virtualNetworks.beginCreateOrUpdateAndWait(
        resourceGroupName,
        vnetName,
        vnetParams
    );

    console.log('Virtual network created');

    return vnet;
}

/**
 * Create Network Interface
 * @param {NetworkManagementClient} networkClient - Azure Network Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region
 * @param {string} vmName - VM name (used for resource naming)
 * @param {object} vnet - Virtual Network object
 * @param {object} publicIP - Public IP object
 * @returns {Promise<object>} Network Interface object
 */
async function createNetworkInterface(networkClient, resourceGroupName, location, vmName, vnet, publicIP) {
    const nicName = `${vmName}-nic`;

    console.log('Creating network interface...');
    const nicParams = {
        location: location,
        ipConfigurations: [{
            name: 'ipconfig1',
            subnet: {
                id: vnet.subnets[0].id
            },
            publicIPAddress: {
                id: publicIP.id
            }
        }]
    };

    const nic = await networkClient.networkInterfaces.beginCreateOrUpdateAndWait(
        resourceGroupName,
        nicName,
        nicParams
    );

    console.log('Network interface created');

    return nic;
}

/**
 * Create Virtual Machine
 * @param {ComputeManagementClient} computeClient - Azure Compute Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region
 * @param {string} vmName - VM name
 * @param {object} nic - Network Interface object
 * @returns {Promise<object>} Virtual Machine object
 */

async function createVirtualMachine(computeClient, resourceGroupName, location, vmName, nic) {
    console.log('Creating virtual machine...');

    const vmParams = {
        location: location,
        hardwareProfile: {
            vmSize: 'Standard_B1s'
        },
        storageProfile: {
            imageReference: {
                publisher: 'Canonical',
                offer: '0001-com-ubuntu-server-jammy',
                sku: '22_04-lts-gen2',
                version: 'latest'
            },
            osDisk: {
                createOption: 'FromImage',
                managedDisk: {
                    storageAccountType: 'Standard_LRS'
                }
            }
        },
        osProfile: {
            computerName: vmName,
            adminUsername: 'azureuser',
            adminPassword: 'Azure123456!'
        },
        networkProfile: {
            networkInterfaces: [{
                id: nic.id,
                primary: true
            }]
        }
    };

    const vm = await computeClient.virtualMachines.beginCreateOrUpdateAndWait(
        resourceGroupName,
        vmName,
        vmParams
    );

    console.log(`Virtual machine created: ${vm.name}`);

    return vm;
}

/**
 * VM Creation - Main Function
 * @param {string} vmId - VM Unique ID
 * @param {string} subscriptionId - Azure Subscription ID
 * @returns {Promise<object>} VM metadata
 *
 * TODO (Phase 3):
 * - Public IP allocation/assignment
 * - noVNC installation (cloud-init)
 */
export async function provisionVM(vmId, subscriptionId) {
    console.log(`Creating VM: ${vmId}`);
    console.log(`Using subscription: ${subscriptionId}`);

    // Azure authentication
    const credential = new DefaultAzureCredential();
    const computeClient = new ComputeManagementClient(
        credential,
        subscriptionId
    );
    const resourceClient = new ResourceManagementClient(
        credential,
        subscriptionId
    );

    // Basic configuration
    const resourceGroupName = process.env.RESOURCE_GROUP_NAME;
    const location = process.env.AZURE_LOCATION;


    try {
        // Step 1: Ensure Resource Group exists
        await ensureResourceGroup(resourceClient, resourceGroupName, location);
        console.log('Resource group ready');

        // Step 2: Create network resources
        const networkClient = new NetworkManagementClient(credential, subscriptionId);
        const vmName = vmId;
        const vnet = await createVirtualNetwork(networkClient, resourceGroupName, location, vmName);
        const publicIP = await createPublicIP(networkClient, resourceGroupName, location, vmName);
        const nic = await createNetworkInterface(networkClient, resourceGroupName, location, vmName, vnet, publicIP);
        console.log('Network resources ready');

        // Step 3: Create VM
        const vm = await createVirtualMachine(computeClient, resourceGroupName, location, vmName, nic);

        console.log('VM creation completed successfully');

        // Verify VM object data
        console.log(`VM ID: ${vm.id}`);
        console.log(`VM Name: ${vm.name}`);
        console.log(`VM Provisioning State: ${vm.provisioningState}`);

        // Refresh NIC to get actual assigned IPs
        const nicDetails = await networkClient.networkInterfaces.get(
            resourceGroupName,
            `${vmName}-nic`
        );

        console.log(`Private IP: ${nicDetails.ipConfigurations[0].privateIPAddress}`);
        console.log(`Public IP ID: ${nicDetails.ipConfigurations[0].publicIPAddress?.id}`);

        // Refresh Public IP to get actual IP address
        const publicIPDetails = await networkClient.publicIPAddresses.get(
            resourceGroupName,
            `${vmName}-ip`
        );

        console.log(`Public IP Address: ${publicIPDetails.ipAddress}`);
        console.log(`Public IP Provisioning State: ${publicIPDetails.provisioningState}`);

        // TODO (Phase 6): Return metadata for GitHub integration
        // return {
        //     vmId: vmId,
        //     vmName: vm.name,
        //     publicIP: publicIPDetails.ipAddress,
        //     privateIP: nicDetails.ipConfigurations[0].privateIPAddress,
        //     resourceGroup: resourceGroupName
        // };

    } catch(error) {
        console.error('VM creation error', error);
        throw error;
    }
}