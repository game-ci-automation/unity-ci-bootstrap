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
 * Create network resources (VNet, Subnet, NIC)
 * @param {NetworkManagementClient} networkClient - Azure Network Management client
 * @param {string} resourceGroupName - Name of the resource group
 * @param {string} location - Azure region
 * @param {string} vmName - VM name (used for resource naming)
 * @returns {Promise<object>} Network Interface object
 */
async function createNetworkResources(networkClient, resourceGroupName, location, vmName) {
    const vnetName = `${vmName}-vnet`;
    const subnetName = `${vmName}-subnet`;
    const nicName = `${vmName}-nic`;

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

    console.log('Creating network interface...');
    const nicParams = {
        location: location,
        ipConfigurations: [{
            name: 'ipconfig1',
            subnet: {
                id: vnet.subnets[0].id
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
        const nic = await createNetworkResources(networkClient, resourceGroupName, location, vmName);
        console.log('Network resources ready');

        // Step 3: Create VM
        const vm = await createVirtualMachine(computeClient, resourceGroupName, location, vmName, nic);

        console.log('VM creation completed successfully');

        return {
            vmId: vmId,
            status: 'succeeded',
            vmName: vm.name,
            resourceGroup: resourceGroupName,
            location: location,
            privateIP: nic.ipConfigurations[0].privateIPAddress
        };

    } catch(error) {
        console.error('VM creation error', error);
        throw error;
    }
}